package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/venue-master/platform/internal/server"
)

type paymentIntentRequest struct {
	AmountCents int64             `json:"amountCents" binding:"required"`
	Currency    string            `json:"currency" binding:"required"`
	Metadata    map[string]string `json:"metadata"`
}

type refundRequest struct {
	PaymentID string `json:"paymentId" binding:"required"`
	Amount    int64  `json:"amountCents"`
}

func main() {
	srv, err := server.New("payment-service")
	if err != nil {
		panic(err)
	}

	registerRoutes(srv.Engine)

	if err := srv.Run(); err != nil {
		panic(err)
	}
}

func registerRoutes(router *gin.Engine) {
	router.POST("/v1/payments/intents", func(ctx *gin.Context) {
		var req paymentIntentRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusCreated, gin.H{
			"id":           "pi_test_123",
			"clientSecret": "secret",
			"status":       "requires_confirmation",
			"amount":       req.AmountCents,
			"currency":     req.Currency,
			"metadata":     req.Metadata,
		})
	})

	router.POST("/v1/payments/refunds", func(ctx *gin.Context) {
		var req refundRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"id":        "re_test_123",
			"paymentId": req.PaymentID,
			"amount":    req.Amount,
			"status":    "succeeded",
			"created":   time.Now().Unix(),
		})
	})
}
