package main

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/venue-master/platform/internal/server"
)

type cartItemRequest struct {
	ProductID string `json:"productId" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required"`
}

func main() {
	srv, err := server.New("shop-service")
	if err != nil {
		panic(err)
	}

	registerRoutes(srv.Engine)

	if err := srv.Run(); err != nil {
		panic(err)
	}
}

func registerRoutes(router *gin.Engine) {
	router.GET("/v1/shop/products", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, []gin.H{
			{"id": "sku-1", "name": "Carbon Paddle", "priceCents": 19900, "inventory": 12},
			{"id": "sku-2", "name": "Pro Balls (pack of 3)", "priceCents": 1200, "inventory": 42},
		})
	})

	router.POST("/v1/shop/cart", func(ctx *gin.Context) {
		var req cartItemRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusCreated, gin.H{"cartId": "cart-1", "items": []gin.H{{"productId": req.ProductID, "quantity": req.Quantity}}})
	})
}
