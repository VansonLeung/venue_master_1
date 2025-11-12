package main

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/venue-master/platform/internal/server"
)

type menuItemRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	PriceCents  int64  `json:"priceCents"`
	Category    string `json:"category"`
	Available   bool   `json:"available"`
}

func main() {
	srv, err := server.New("food-service")
	if err != nil {
		panic(err)
	}

	registerRoutes(srv.Engine)

	if err := srv.Run(); err != nil {
		panic(err)
	}
}

func registerRoutes(router *gin.Engine) {
	router.GET("/v1/menu", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"items": []gin.H{
				{
					"id":          "menu-1",
					"name":        "Signature Smoothie",
					"description": "Kale, pineapple, chia seeds",
					"priceCents":  1200,
					"category":    "Beverages",
					"available":   true,
				},
			},
		})
	})

	router.POST("/v1/menu/items", func(ctx *gin.Context) {
		var req menuItemRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusCreated, gin.H{"id": "menu-2"})
	})

	router.PATCH("/v1/menu/items/:id/availability", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"id": ctx.Param("id"), "available": ctx.Query("value") == "true"})
	})
}
