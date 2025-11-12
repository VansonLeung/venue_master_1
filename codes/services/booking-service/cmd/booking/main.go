package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"

	"github.com/venue-master/platform/internal/server"
	"github.com/venue-master/platform/lib/config"
	"github.com/venue-master/platform/services/booking-service/internal/payment"
	"github.com/venue-master/platform/services/booking-service/internal/store"
)

const (
	defaultVenueID  = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	defaultFacility = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
)

type handler struct {
	store   *store.Store
	payment *payment.Client
	logger  zerolog.Logger
}

func main() {
	srv, err := server.New("booking-service")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	repo, err := initStore(ctx, srv.Config.Database, srv.Logger)
	if err != nil {
		panic(err)
	}
	defer repo.Close()

	if err := seedDefaultFacility(ctx, repo, srv.Logger); err != nil {
		panic(err)
	}

	paymentClient := payment.New(getEnv("PAYMENT_SERVICE_URL", "http://payment-service:8080"))
	h := &handler{store: repo, payment: paymentClient, logger: srv.Logger}
	registerRoutes(srv.Engine, h)

	if err := srv.Run(); err != nil {
		panic(err)
	}
}

func registerRoutes(router *gin.Engine, h *handler) {
	router.GET("/v1/bookings", h.listBookings)
	router.GET("/v1/bookings/:id", h.getBooking)
	router.POST("/v1/bookings", h.createBooking)
	router.DELETE("/v1/bookings/:id", h.cancelBooking)

	router.GET("/v1/facilities", h.listFacilities)
	router.POST("/v1/facilities", h.createFacility)
	router.PATCH("/v1/facilities/:id", h.updateFacilityAvailability)
}

type bookingRequest struct {
	FacilityID string `json:"facilityId" binding:"required"`
	UserID     string `json:"userId" binding:"required"`
	StartsAt   string `json:"startsAt" binding:"required"`
	EndsAt     string `json:"endsAt" binding:"required"`
}

type facilityRequest struct {
	VenueID     string `json:"venueId" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Surface     string `json:"surface"`
	OpenAt      string `json:"openAt" binding:"required"`
	CloseAt     string `json:"closeAt" binding:"required"`
}

type availabilityRequest struct {
	Available bool `json:"available"`
}

func (h *handler) listBookings(ctx *gin.Context) {
	userIDParam := ctx.Query("userId")
	var userID uuid.UUID
	var err error
	if userIDParam != "" {
		userID, err = uuid.Parse(userIDParam)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid userId"})
			return
		}
	}

	bookings, err := h.store.ListBookings(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, bookingsResponse(bookings))
}

func (h *handler) getBooking(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking id"})
		return
	}
	booking, err := h.store.GetBooking(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, bookingResponse(*booking))
}

func (h *handler) createBooking(ctx *gin.Context) {
	var req bookingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	facilityID, err := uuid.Parse(req.FacilityID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid facilityId"})
		return
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid userId"})
		return
	}
	startsAt, err := time.Parse(time.RFC3339, req.StartsAt)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid startsAt"})
		return
	}
	endsAt, err := time.Parse(time.RFC3339, req.EndsAt)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid endsAt"})
		return
	}

	booking, err := h.store.CreateBooking(ctx, store.CreateBookingInput{
		FacilityID:  facilityID,
		UserID:      userID,
		StartsAt:    startsAt,
		EndsAt:      endsAt,
		AmountCents: 4500,
		Currency:    "CAD",
	})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	intent, err := h.payment.Charge(ctx, booking.AmountCents, booking.Currency, map[string]string{
		"booking_id":  booking.ID.String(),
		"facility_id": booking.FacilityID.String(),
	})
	if err != nil {
		h.logger.Warn().Err(err).Str("booking_id", booking.ID.String()).Msg("payment intent failed")
		booking, _ = h.store.UpdateBookingStatus(ctx, booking.ID, "PAYMENT_FAILED", "")
		ctx.JSON(http.StatusBadGateway, gin.H{"error": "payment failed"})
		return
	}

	booking, err = h.store.UpdateBookingStatus(ctx, booking.ID, "CONFIRMED", intent.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := h.store.AttachFacility(ctx, booking); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, bookingResponse(*booking))
}

func (h *handler) cancelBooking(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking id"})
		return
	}
	booking, err := h.store.CancelBooking(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := h.store.AttachFacility(ctx, booking); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, bookingResponse(*booking))
}

func (h *handler) listFacilities(ctx *gin.Context) {
	venueParam := ctx.Query("venueId")
	var venueID uuid.UUID
	var err error
	if venueParam != "" {
		venueID, err = uuid.Parse(venueParam)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid venueId"})
			return
		}
	}
	onlyAvailable := ctx.Query("available") == "true"

	facilities, err := h.store.ListFacilities(ctx, venueID, onlyAvailable)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, facilitiesResponse(facilities))
}

func (h *handler) createFacility(ctx *gin.Context) {
	var req facilityRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	venueID, err := uuid.Parse(req.VenueID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid venueId"})
		return
	}
	openAt, err := time.Parse("15:04", req.OpenAt)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid openAt, use HH:MM"})
		return
	}
	closeAt, err := time.Parse("15:04", req.CloseAt)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid closeAt, use HH:MM"})
		return
	}

	facility := store.Facility{
		ID:          uuid.New(),
		VenueID:     venueID,
		Name:        req.Name,
		Description: req.Description,
		Surface:     req.Surface,
		OpenAt:      openAt,
		CloseAt:     closeAt,
		Available:   true,
	}
	created, err := h.store.CreateFacility(ctx, facility)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, facilityResponse(*created))
}

func (h *handler) updateFacilityAvailability(ctx *gin.Context) {
	facilityID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid facility id"})
		return
	}
	var req availabilityRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	facility, err := h.store.UpdateFacilityAvailability(ctx, facilityID, req.Available)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, facilityResponse(*facility))
}

func seedDefaultFacility(ctx context.Context, repo *store.Store, logger zerolog.Logger) error {
	venueID := uuid.MustParse(defaultVenueID)
	facilityID := uuid.MustParse(defaultFacility)
	openAt, _ := time.Parse("15:04", "06:00")
	closeAt, _ := time.Parse("15:04", "23:00")
	facility := store.Facility{
		ID:          facilityID,
		VenueID:     venueID,
		Name:        "Center Court",
		Description: "Indoor pickleball court",
		Surface:     "hardwood",
		OpenAt:      openAt,
		CloseAt:     closeAt,
		Available:   true,
	}
	if err := repo.SeedFacility(ctx, facility); err != nil {
		return err
	}
	logger.Info().Msg("default facility ensured")
	return nil
}

func initStore(ctx context.Context, cfg config.DatabaseConfig, logger zerolog.Logger) (*store.Store, error) {
	deadline := time.Now().Add(60 * time.Second)
	var lastErr error
	for {
		repo, err := store.New(ctx, cfg)
		if err == nil {
			if err = repo.RunMigrations(ctx); err == nil {
				logger.Info().Msg("booking store ready")
				return repo, nil
			}
			repo.Close()
			lastErr = err
		} else {
			lastErr = err
		}
		if time.Now().After(deadline) {
			break
		}
		logger.Warn().Err(lastErr).Msg("waiting for postgres")
		time.Sleep(3 * time.Second)
	}
	return nil, lastErr
}

func bookingResponse(b store.Booking) gin.H {
	resp := gin.H{
		"id":            b.ID,
		"facilityId":    b.FacilityID,
		"userId":        b.UserID,
		"startsAt":      b.StartsAt.Format(time.RFC3339),
		"endsAt":        b.EndsAt.Format(time.RFC3339),
		"status":        b.Status,
		"amountCents":   b.AmountCents,
		"currency":      b.Currency,
		"paymentIntent": b.PaymentIntent,
	}
	if b.Facility != nil {
		resp["facility"] = facilityResponse(*b.Facility)
	}
	return resp
}

func bookingsResponse(items []store.Booking) []gin.H {
	out := make([]gin.H, 0, len(items))
	for _, b := range items {
		out = append(out, bookingResponse(b))
	}
	return out
}

func facilityResponse(f store.Facility) gin.H {
	return gin.H{
		"id":          f.ID,
		"venueId":     f.VenueID,
		"name":        f.Name,
		"description": f.Description,
		"surface":     f.Surface,
		"openAt":      f.OpenAt.Format("15:04"),
		"closeAt":     f.CloseAt.Format("15:04"),
		"available":   f.Available,
	}
}

func facilitiesResponse(items []store.Facility) []gin.H {
	out := make([]gin.H, 0, len(items))
	for _, f := range items {
		out = append(out, facilityResponse(f))
	}
	return out
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}
