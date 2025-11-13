package main

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"

	"github.com/venue-master/platform/internal/server"
	"github.com/venue-master/platform/lib/config"
	"github.com/venue-master/platform/services/booking-service/internal/middleware"
	"github.com/venue-master/platform/services/booking-service/internal/notification"
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
	notify  *notification.Client
	logger  zerolog.Logger
}

const paymentRetryMaxAttempts = 5

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
	notificationClient := notification.New(getEnv("NOTIFICATION_SERVICE_URL", "http://notification-service:8080"))
	h := &handler{store: repo, payment: paymentClient, notify: notificationClient, logger: srv.Logger}
	registerRoutes(srv.Engine, h)

	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.startRetryWorker(appCtx)

	if err := srv.Run(); err != nil {
		panic(err)
	}
}

func registerRoutes(router *gin.Engine, h *handler) {
	router.Use(middleware.RequireAuth())

	readRoles := []string{middleware.RoleMember, middleware.RoleOperator, middleware.RoleAdmin, middleware.RoleVenueAdmin}
	memberWriteRoles := []string{middleware.RoleMember, middleware.RoleAdmin, middleware.RoleVenueAdmin}
	adminRoles := []string{middleware.RoleAdmin, middleware.RoleVenueAdmin}

	router.GET("/v1/bookings", middleware.RequireRoles(readRoles...), h.listBookings)
	router.GET("/v1/bookings/:id", middleware.RequireRoles(readRoles...), h.getBooking)
	router.POST("/v1/bookings", middleware.RequireRoles(memberWriteRoles...), h.createBooking)
	router.DELETE("/v1/bookings/:id", middleware.RequireRoles(memberWriteRoles...), h.cancelBooking)

	router.GET("/v1/facilities", middleware.RequireRoles(readRoles...), h.listFacilities)
	router.POST("/v1/facilities", middleware.RequireRoles(adminRoles...), h.createFacility)
	router.PATCH("/v1/facilities/:id", middleware.RequireRoles(adminRoles...), h.updateFacilityAvailability)
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
	WeekdayRate int    `json:"weekdayRateCents"`
	WeekendRate int    `json:"weekendRateCents"`
	Currency    string `json:"currency"`
}

type availabilityRequest struct {
	Available bool `json:"available"`
}

func (h *handler) listBookings(ctx *gin.Context) {
	user, _ := middleware.GetUser(ctx)
	var filter uuid.UUID

	if userIDParam := ctx.Query("userId"); userIDParam != "" {
		id, ok := uuidFromString(ctx, userIDParam, "userId")
		if !ok {
			return
		}
		if !isAdmin(user) && id.String() != user.UserID {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		filter = id
	} else if !isAdmin(user) {
		id, ok := uuidFromString(ctx, user.UserID, "userId")
		if !ok {
			return
		}
		filter = id
	}

	limit, offset, ok := paginationParams(ctx)
	if !ok {
		return
	}
	bookings, err := h.store.ListBookings(ctx, filter, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, bookingsResponse(bookings))
}

func (h *handler) getBooking(ctx *gin.Context) {
	user, _ := middleware.GetUser(ctx)
	id, ok := uuidFromString(ctx, ctx.Param("id"), "booking id")
	if !ok {
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
	if !isAdmin(user) && booking.UserID.String() != user.UserID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	ctx.JSON(http.StatusOK, bookingResponse(*booking))
}

func (h *handler) createBooking(ctx *gin.Context) {
	user, _ := middleware.GetUser(ctx)
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
	facility, err := h.store.GetFacility(ctx, facilityID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "facility not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !facility.Available && !isAdmin(user) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "facility unavailable"})
		return
	}
	if req.UserID == "" {
		req.UserID = user.UserID
	}
	if !isAdmin(user) && req.UserID != user.UserID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	userID, ok := uuidFromString(ctx, req.UserID, "userId")
	if !ok {
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

	amount := calculateBookingAmount(facility, startsAt, endsAt)
	booking, err := h.store.CreateBooking(ctx, store.CreateBookingInput{
		FacilityID:  facilityID,
		UserID:      userID,
		StartsAt:    startsAt,
		EndsAt:      endsAt,
		AmountCents: amount,
		Currency:    facility.Currency,
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
		if _, updateErr := h.store.UpdateBookingStatus(ctx, booking.ID, "PAYMENT_RETRY", ""); updateErr != nil {
			h.logger.Error().Err(updateErr).Str("booking_id", booking.ID.String()).Msg("failed to mark booking retry")
		}
		h.schedulePaymentRetry(ctx, booking.ID, err)
		if attachErr := h.store.AttachFacility(ctx, booking); attachErr == nil {
			ctx.JSON(http.StatusAccepted, bookingResponse(*booking))
		} else {
			ctx.JSON(http.StatusAccepted, gin.H{"status": "PAYMENT_RETRY"})
		}
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
	user, _ := middleware.GetUser(ctx)
	id, ok := uuidFromString(ctx, ctx.Param("id"), "booking id")
	if !ok {
		return
	}
	existing, err := h.store.GetBooking(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !isAdmin(user) && existing.UserID.String() != user.UserID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
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
	limit, offset, ok := paginationParams(ctx)
	if !ok {
		return
	}
	var availablePtr *bool
	if val := ctx.Query("available"); val != "" {
		switch strings.ToLower(val) {
		case "true", "1":
			b := true
			availablePtr = &b
		case "false", "0":
			b := false
			availablePtr = &b
		default:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid available value"})
			return
		}
	}

	facilities, err := h.store.ListFacilities(ctx, venueID, availablePtr, limit, offset)
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
	weekdayRate := req.WeekdayRate
	if weekdayRate <= 0 {
		weekdayRate = 4500
	}
	weekendRate := req.WeekendRate
	if weekendRate <= 0 {
		weekendRate = weekdayRate
	}
	currency := req.Currency
	if currency == "" {
		currency = "CAD"
	}

	facility := store.Facility{
		ID:               uuid.New(),
		VenueID:          venueID,
		Name:             req.Name,
		Description:      req.Description,
		Surface:          req.Surface,
		OpenAt:           openAt,
		CloseAt:          closeAt,
		Available:        true,
		WeekdayRateCents: weekdayRate,
		WeekendRateCents: weekendRate,
		Currency:         currency,
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
		"id":               f.ID,
		"venueId":          f.VenueID,
		"name":             f.Name,
		"description":      f.Description,
		"surface":          f.Surface,
		"openAt":           f.OpenAt.Format("15:04"),
		"closeAt":          f.CloseAt.Format("15:04"),
		"available":        f.Available,
		"weekdayRateCents": f.WeekdayRateCents,
		"weekendRateCents": f.WeekendRateCents,
		"currency":         f.Currency,
	}
}

func facilitiesResponse(items []store.Facility) []gin.H {
	out := make([]gin.H, 0, len(items))
	for _, f := range items {
		out = append(out, facilityResponse(f))
	}
	return out
}

func uuidFromString(ctx *gin.Context, value, field string) (uuid.UUID, bool) {
	id, err := uuid.Parse(value)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid %s", field)})
		return uuid.UUID{}, false
	}
	return id, true
}

func isAdmin(user middleware.ContextUser) bool {
	return user.HasAnyRole(middleware.RoleAdmin, middleware.RoleVenueAdmin)
}

func paginationParams(ctx *gin.Context) (int, int, bool) {
	limit := 20
	offset := 0
	if raw := ctx.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
			return 0, 0, false
		}
		limit = parsed
	}
	if limit > 100 {
		limit = 100
	}
	if raw := ctx.Query("offset"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset"})
			return 0, 0, false
		}
		offset = parsed
	}
	return limit, offset, true
}

func (h *handler) schedulePaymentRetry(ctx context.Context, bookingID uuid.UUID, cause error) {
	errMsg := ""
	if cause != nil {
		errMsg = cause.Error()
	}
	next := time.Now().Add(time.Minute)
	if err := h.store.SchedulePaymentRetry(ctx, bookingID, next, 1, errMsg); err != nil {
		h.logger.Error().Err(err).Str("booking_id", bookingID.String()).Msg("failed to schedule payment retry")
	}
}

func (h *handler) startRetryWorker(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.processPaymentRetries(ctx)
		}
	}
}

func (h *handler) processPaymentRetries(ctx context.Context) {
	retries, err := h.store.FetchDuePaymentRetries(ctx, 10)
	if err != nil {
		h.logger.Error().Err(err).Msg("fetch payment retries failed")
		return
	}
	for _, retry := range retries {
		h.handleSingleRetry(ctx, retry)
	}
}

func (h *handler) handleSingleRetry(ctx context.Context, retry store.PaymentRetry) {
	attempt := retry.Attempt
	if attempt > paymentRetryMaxAttempts {
		attempt = paymentRetryMaxAttempts
	}
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	booking, err := h.store.GetBooking(ctxTimeout, retry.BookingID)
	if err != nil {
		h.logger.Error().Err(err).Str("booking_id", retry.BookingID.String()).Msg("failed to load booking for retry")
		_ = h.store.DeletePaymentRetry(ctx, retry.BookingID)
		return
	}

	metadata := map[string]string{
		"booking_id":    booking.ID.String(),
		"facility_id":   booking.FacilityID.String(),
		"retry_attempt": fmt.Sprintf("%d", attempt),
	}
	intent, err := h.payment.Charge(ctxTimeout, booking.AmountCents, booking.Currency, metadata)
	if err == nil {
		if _, updateErr := h.store.UpdateBookingStatus(ctxTimeout, booking.ID, "CONFIRMED", intent.ID); updateErr != nil {
			h.logger.Error().Err(updateErr).Str("booking_id", booking.ID.String()).Msg("failed to update booking after retry success")
		}
		_ = h.store.DeletePaymentRetry(ctxTimeout, booking.ID)
		return
	}

	nextAttempt := attempt + 1
	if nextAttempt > paymentRetryMaxAttempts {
		if _, updateErr := h.store.UpdateBookingStatus(ctxTimeout, booking.ID, "PAYMENT_FAILED", ""); updateErr != nil {
			h.logger.Error().Err(updateErr).Str("booking_id", booking.ID.String()).Msg("failed to mark booking failed")
		}
		_ = h.store.DeletePaymentRetry(ctxTimeout, booking.ID)
		h.notifyFailure(ctxTimeout, booking, err)
		return
	}

	delay := time.Minute * time.Duration(1<<uint(nextAttempt-1))
	next := time.Now().Add(delay)
	if schedErr := h.store.SchedulePaymentRetry(ctxTimeout, booking.ID, next, nextAttempt, err.Error()); schedErr != nil {
		h.logger.Error().Err(schedErr).Str("booking_id", booking.ID.String()).Msg("failed to reschedule payment retry")
	}
}

func (h *handler) notifyFailure(ctx context.Context, booking *store.Booking, cause error) {
	if h.notify == nil || booking == nil {
		return
	}
	message := fmt.Sprintf("We were unable to process payment for booking %s after multiple attempts.", booking.ID)
	if cause != nil {
		message = fmt.Sprintf("%s Error: %s", message, cause.Error())
	}
	payload := notification.NotifyPayload{
		UserID:  booking.UserID.String(),
		Title:   "Payment Failed",
		Message: message,
		Channel: "in_app",
	}
	if err := h.notify.Send(ctx, payload); err != nil {
		h.logger.Error().Err(err).Msg("failed to send payment failure notification")
	}
}
func calculateBookingAmount(f *store.Facility, start, end time.Time) int {
	duration := end.Sub(start)
	if duration <= 0 {
		duration = time.Hour
	}
	hours := int(math.Ceil(duration.Hours()))
	rate := f.WeekdayRateCents
	if isWeekend(start) {
		rate = f.WeekendRateCents
	}
	return rate * hours
}

func isWeekend(t time.Time) bool {
	switch t.Weekday() {
	case time.Saturday, time.Sunday:
		return true
	default:
		return false
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}
