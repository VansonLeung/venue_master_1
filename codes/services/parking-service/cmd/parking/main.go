package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/venue-master/platform/internal/server"
)

type parkingRequest struct {
	SpotID string `json:"spotId" binding:"required"`
	Start  string `json:"start" binding:"required"`
	End    string `json:"end" binding:"required"`
}

func main() {
	srv, err := server.New("parking-service")
	if err != nil {
		panic(err)
	}

	registerRoutes(srv.Engine)

	if err := srv.Run(); err != nil {
		panic(err)
	}
}

func registerRoutes(router *gin.Engine) {
	router.GET("/v1/parking/spaces", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, []gin.H{
			{"id": "spot-1", "label": "P1", "occupied": false},
			{"id": "spot-2", "label": "P2", "occupied": true},
		})
	})

	router.POST("/v1/parking/reservations", func(ctx *gin.Context) {
		var req parkingRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusCreated, gin.H{
			"id":      "parking-reservation-1",
			"spotId":  req.SpotID,
			"start":   req.Start,
			"end":     req.End,
			"status":  "CONFIRMED",
			"created": time.Now().Format(time.RFC3339),
		})
	})

	router.DELETE("/v1/parking/reservations/:id", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "CANCELLED", "id": ctx.Param("id")})
	})
}
