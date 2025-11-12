package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/venue-master/platform/internal/server"
)

type bookingRequest struct {
	FacilityID string `json:"facilityId" binding:"required"`
	UserID     string `json:"userId" binding:"required"`
	StartsAt   string `json:"startsAt" binding:"required"`
	EndsAt     string `json:"endsAt" binding:"required"`
}

func main() {
	srv, err := server.New("booking-service")
	if err != nil {
		panic(err)
	}

	registerRoutes(srv.Engine)

	if err := srv.Run(); err != nil {
		panic(err)
	}
}

func registerRoutes(router *gin.Engine) {
	router.GET("/v1/bookings", func(ctx *gin.Context) {
		userID := ctx.Query("userId")
		ctx.JSON(http.StatusOK, []gin.H{mockBooking(userID)})
	})

	router.POST("/v1/bookings", func(ctx *gin.Context) {
		var req bookingRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusCreated, bookingResponse(req))
	})

	router.DELETE("/v1/bookings/:id", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, mockBooking(ctx.Query("userId")))
	})

	router.GET("/v1/facilities", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, []gin.H{mockFacility(ctx.Query("venueId"))})
	})
}

func mockBooking(userID string) gin.H {
	if userID == "" {
		userID = "user-1"
	}
	start := time.Now().Add(24 * time.Hour).Truncate(time.Hour)
	return gin.H{
		"id":          "booking-1",
		"facilityId":  "facility-1",
		"userId":      userID,
		"startsAt":    start.Format(time.RFC3339),
		"endsAt":      start.Add(90 * time.Minute).Format(time.RFC3339),
		"status":      "CONFIRMED",
		"amountCents": 4500,
		"currency":    "CAD",
	}
}

func bookingResponse(req bookingRequest) gin.H {
	return gin.H{
		"id":          "booking-" + req.FacilityID,
		"facilityId":  req.FacilityID,
		"userId":      req.UserID,
		"startsAt":    req.StartsAt,
		"endsAt":      req.EndsAt,
		"status":      "PENDING_PAYMENT",
		"amountCents": 4500,
		"currency":    "CAD",
	}
}

func mockFacility(venueID string) gin.H {
	if venueID == "" {
		venueID = "venue-1"
	}
	open := time.Date(0, 1, 1, 6, 0, 0, 0, time.UTC)
	close := time.Date(0, 1, 1, 23, 0, 0, 0, time.UTC)
	return gin.H{
		"id":          "facility-1",
		"venueId":     venueID,
		"name":        "Center Court",
		"description": "Indoor pickleball court",
		"surface":     "hardwood",
		"openAt":      open.Format(time.RFC3339),
		"closeAt":     close.Format(time.RFC3339),
	}
}
