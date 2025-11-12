package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/venue-master/platform/internal/server"
)

type notificationRequest struct {
	UserID  string `json:"userId" binding:"required"`
	Title   string `json:"title" binding:"required"`
	Message string `json:"message" binding:"required"`
	Channel string `json:"channel" binding:"required"`
}

func main() {
	srv, err := server.New("notification-service")
	if err != nil {
		panic(err)
	}

	registerRoutes(srv.Engine)

	if err := srv.Run(); err != nil {
		panic(err)
	}
}

func registerRoutes(router *gin.Engine) {
	router.GET("/v1/notifications", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, []gin.H{
			{"id": "notif-1", "title": "Booking Confirmed", "message": "See you tomorrow!", "channel": "push", "read": false},
		})
	})

	router.POST("/v1/notifications", func(ctx *gin.Context) {
		var req notificationRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusAccepted, gin.H{
			"id":        "notif-" + req.Channel,
			"userId":    req.UserID,
			"title":     req.Title,
			"message":   req.Message,
			"channel":   req.Channel,
			"createdAt": time.Now().Format(time.RFC3339),
		})
	})
}
