package rest

import (
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/venue-master/platform/lib/jwtutil"
	"github.com/venue-master/platform/services/api-gateway/internal/services"
)

type Handler struct {
	clients     *services.ServiceClients
	jwtManager  *jwtutil.Manager
	logger      zerolog.Logger
	bookingURL  string
	userURL     string
}

func New(clients *services.ServiceClients, jwtManager *jwtutil.Manager, logger zerolog.Logger) *Handler {
	bookingURL := os.Getenv("BOOKING_SERVICE_URL")
	if bookingURL == "" {
		bookingURL = "http://booking-service:8080"
	}
	userURL := os.Getenv("USER_SERVICE_URL")
	if userURL == "" {
		userURL = "http://user-service:8080"
	}

	return &Handler{
		clients:     clients,
		jwtManager:  jwtManager,
		logger:      logger,
		bookingURL:  strings.TrimRight(bookingURL, "/"),
		userURL:     strings.TrimRight(userURL, "/"),
	}
}

// Register adds REST endpoints to the gin engine
func (h *Handler) Register(engine *gin.Engine) {
	// Auth middleware for REST endpoints
	authMiddleware := h.authMiddleware()

	// Venues endpoints - proxy to booking service
	venues := engine.Group("/v1/venues", authMiddleware)
	{
		venues.GET("", h.listVenues)
		venues.GET("/:id", h.getVenue)
		venues.POST("", h.createVenue)
		venues.PUT("/:id", h.updateVenue)
		venues.DELETE("/:id", h.deleteVenue)
	}

	// Facilities endpoints - proxy to booking service
	facilities := engine.Group("/v1/facilities", authMiddleware)
	{
		facilities.GET("", h.listFacilities)
		facilities.GET("/:id", h.getFacility)
		facilities.POST("", h.createFacility)
		facilities.PUT("/:id", h.updateFacility)
		facilities.DELETE("/:id", h.deleteFacility)
		facilities.GET("/:id/schedule", h.getFacilitySchedule)
	}

	// Bookings endpoints - proxy to booking service
	bookings := engine.Group("/v1/bookings", authMiddleware)
	{
		bookings.GET("", h.listBookings)
		bookings.GET("/:id", h.getBooking)
		bookings.POST("", h.createBooking)
		bookings.PATCH("/:id/status", h.updateBookingStatus)
		bookings.PATCH("/:id/cancel", h.cancelBooking)
		bookings.POST("/:id/confirm", h.confirmBooking)
		bookings.GET("/stats", h.getBookingStats)
	}

	// Users endpoints - proxy to user service
	users := engine.Group("/v1/users", authMiddleware)
	{
		users.GET("", h.listUsers)
		users.GET("/:id", h.getUser)
	}
}

// authMiddleware validates JWT and injects user context
func (h *Handler) authMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			ctx.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			ctx.Abort()
			return
		}

		claims, err := h.jwtManager.Validate(token)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			ctx.Abort()
			return
		}

		// Store auth metadata in context for downstream requests
		authCtx := services.WithAuth(ctx.Request.Context(), services.AuthMetadata{
			UserID: claims.UserID,
			Roles:  claims.Roles,
		})
		ctx.Request = ctx.Request.WithContext(authCtx)

		ctx.Next()
	}
}

// proxyRequest forwards the request to the booking service
func (h *Handler) proxyRequest(ctx *gin.Context, targetURL, method, path string, body io.Reader) {
	fullURL := targetURL + path

	req, err := http.NewRequestWithContext(ctx.Request.Context(), method, fullURL, body)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to create proxy request")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
		return
	}

	// Copy headers
	req.Header.Set("Content-Type", "application/json")

	// Inject auth headers from context
	if meta, ok := services.AuthFromContext(ctx.Request.Context()); ok {
		if meta.UserID != "" {
			req.Header.Set("X-User-ID", meta.UserID)
		}
		if len(meta.Roles) > 0 {
			req.Header.Set("X-User-Roles", strings.Join(meta.Roles, ","))
		}
	}

	// Forward request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.Error().Err(err).Str("url", fullURL).Msg("failed to forward request")
		ctx.JSON(http.StatusBadGateway, gin.H{"error": "failed to forward request"})
		return
	}
	defer resp.Body.Close()

	// Copy response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to read response")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read response"})
		return
	}

	// Forward status and response
	for key, values := range resp.Header {
		for _, value := range values {
			ctx.Header(key, value)
		}
	}
	ctx.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

// Facility handlers
func (h *Handler) listFacilities(ctx *gin.Context) {
	path := "/v1/facilities?" + ctx.Request.URL.RawQuery
	h.proxyRequest(ctx, h.bookingURL, http.MethodGet, path, nil)
}

func (h *Handler) getFacility(ctx *gin.Context) {
	path := "/v1/facilities/" + ctx.Param("id")
	h.proxyRequest(ctx, h.bookingURL, http.MethodGet, path, nil)
}

func (h *Handler) createFacility(ctx *gin.Context) {
	path := "/v1/facilities"
	h.proxyRequest(ctx, h.bookingURL, http.MethodPost, path, ctx.Request.Body)
}

func (h *Handler) updateFacility(ctx *gin.Context) {
	path := "/v1/facilities/" + ctx.Param("id")
	h.proxyRequest(ctx, h.bookingURL, http.MethodPut, path, ctx.Request.Body)
}

func (h *Handler) deleteFacility(ctx *gin.Context) {
	path := "/v1/facilities/" + ctx.Param("id")
	h.proxyRequest(ctx, h.bookingURL, http.MethodDelete, path, nil)
}

func (h *Handler) getFacilitySchedule(ctx *gin.Context) {
	path := "/v1/facilities/" + ctx.Param("id") + "/schedule?" + ctx.Request.URL.RawQuery
	h.proxyRequest(ctx, h.bookingURL, http.MethodGet, path, nil)
}

// Booking handlers
func (h *Handler) listBookings(ctx *gin.Context) {
	path := "/v1/bookings?" + ctx.Request.URL.RawQuery
	h.proxyRequest(ctx, h.bookingURL, http.MethodGet, path, nil)
}

func (h *Handler) getBooking(ctx *gin.Context) {
	path := "/v1/bookings/" + ctx.Param("id")
	h.proxyRequest(ctx, h.bookingURL, http.MethodGet, path, nil)
}

func (h *Handler) createBooking(ctx *gin.Context) {
	path := "/v1/bookings"
	h.proxyRequest(ctx, h.bookingURL, http.MethodPost, path, ctx.Request.Body)
}

func (h *Handler) updateBookingStatus(ctx *gin.Context) {
	path := "/v1/bookings/" + ctx.Param("id") + "/status"
	h.proxyRequest(ctx, h.bookingURL, http.MethodPatch, path, ctx.Request.Body)
}

func (h *Handler) cancelBooking(ctx *gin.Context) {
	path := "/v1/bookings/" + ctx.Param("id") + "/cancel"
	h.proxyRequest(ctx, h.bookingURL, http.MethodPatch, path, nil)
}

func (h *Handler) confirmBooking(ctx *gin.Context) {
	path := "/v1/bookings/" + ctx.Param("id") + "/confirm"
	h.proxyRequest(ctx, h.bookingURL, http.MethodPost, path, nil)
}

func (h *Handler) getBookingStats(ctx *gin.Context) {
	path := "/v1/bookings/stats?" + ctx.Request.URL.RawQuery
	h.proxyRequest(ctx, h.bookingURL, http.MethodGet, path, nil)
}

// Venue handlers
func (h *Handler) listVenues(ctx *gin.Context) {
	path := "/v1/venues?" + ctx.Request.URL.RawQuery
	h.proxyRequest(ctx, h.bookingURL, http.MethodGet, path, nil)
}

func (h *Handler) getVenue(ctx *gin.Context) {
	path := "/v1/venues/" + ctx.Param("id")
	h.proxyRequest(ctx, h.bookingURL, http.MethodGet, path, nil)
}

func (h *Handler) createVenue(ctx *gin.Context) {
	path := "/v1/venues"
	h.proxyRequest(ctx, h.bookingURL, http.MethodPost, path, ctx.Request.Body)
}

func (h *Handler) updateVenue(ctx *gin.Context) {
	path := "/v1/venues/" + ctx.Param("id")
	h.proxyRequest(ctx, h.bookingURL, http.MethodPut, path, ctx.Request.Body)
}

func (h *Handler) deleteVenue(ctx *gin.Context) {
	path := "/v1/venues/" + ctx.Param("id")
	h.proxyRequest(ctx, h.bookingURL, http.MethodDelete, path, nil)
}

// User handlers
func (h *Handler) listUsers(ctx *gin.Context) {
	// For simplicity, return empty array
	// You can implement full user listing if needed
	ctx.JSON(http.StatusOK, []gin.H{})
}

func (h *Handler) getUser(ctx *gin.Context) {
	userID := ctx.Param("id")
	user, err := h.clients.Users.Me(ctx.Request.Context(), userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":        user.ID,
		"email":     user.Email,
		"firstName": user.FirstName,
		"lastName":  user.LastName,
		"roles":     user.Roles,
	})
}
