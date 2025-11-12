package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/venue-master/platform/internal/server"
	"github.com/venue-master/platform/lib/errutil"
	"github.com/venue-master/platform/lib/jwtutil"
	"github.com/venue-master/platform/services/auth-service/internal/session"
	"github.com/venue-master/platform/services/auth-service/internal/userclient"
)

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

func main() {
	srv, err := server.New("auth-service")
	if err != nil {
		panic(err)
	}

	jwtManager := jwtutil.NewManager(srv.Config.JWT)
	sessions, err := session.NewStore(srv.Config.Redis, jwtManager.RefreshTTL())
	if err != nil {
		panic(err)
	}

	userServiceURL := os.Getenv("USER_SERVICE_URL")
	if userServiceURL == "" {
		userServiceURL = "http://user-service:8080"
	}
	users := userclient.New(userServiceURL)

	registerRoutes(srv.Engine, jwtManager, users, sessions)

	if err := srv.Run(); err != nil {
		panic(err)
	}
}

func registerRoutes(router *gin.Engine, jwtManager *jwtutil.Manager, users *userclient.Client, sessions *session.Store) {
	group := router.Group("/v1/auth")
	group.POST("/login", func(ctx *gin.Context) {
		var req loginRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			errutil.Write(ctx, http.StatusBadRequest, "invalid_credentials", "Email and password are required", err.Error())
			return
		}

		timeoutCtx, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
		defer cancel()

		user, err := users.Authenticate(timeoutCtx, req.Email, req.Password)
		if err != nil {
			errutil.Write(ctx, http.StatusUnauthorized, "invalid_credentials", "Email or password incorrect", nil)
			return
		}

		access, refresh, err := jwtManager.Generate(user.ID, user.Roles, nil)
		if err != nil {
			errutil.HandleInternal(ctx, err)
			return
		}

		if err := sessions.Save(timeoutCtx, user.ID, refresh); err != nil {
			errutil.HandleInternal(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"accessToken":  access,
			"refreshToken": refresh,
			"expiresIn":    int(jwtManager.AccessTTL().Seconds()),
			"user":         user,
		})
	})

	group.POST("/refresh", func(ctx *gin.Context) {
		var req refreshRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			errutil.Write(ctx, http.StatusBadRequest, "invalid_request", "refreshToken is required", err.Error())
			return
		}

		claims, err := jwtManager.Validate(req.RefreshToken)
		if err != nil {
			errutil.Write(ctx, http.StatusUnauthorized, "invalid_refresh_token", "Refresh token invalid", err.Error())
			return
		}

		timeoutCtx, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
		defer cancel()

		exists, err := sessions.Exists(timeoutCtx, claims.UserID, req.RefreshToken)
		if err != nil {
			errutil.HandleInternal(ctx, err)
			return
		}
		if !exists {
			errutil.Write(ctx, http.StatusUnauthorized, "invalid_refresh_token", "Refresh token revoked", nil)
			return
		}

		access, newRefresh, err := jwtManager.Refresh(req.RefreshToken)
		if err != nil {
			errutil.Write(ctx, http.StatusUnauthorized, "invalid_refresh_token", "Refresh token invalid", err.Error())
			return
		}

		if err := sessions.Delete(timeoutCtx, claims.UserID, req.RefreshToken); err != nil {
			errutil.HandleInternal(ctx, err)
			return
		}
		if err := sessions.Save(timeoutCtx, claims.UserID, newRefresh); err != nil {
			errutil.HandleInternal(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"accessToken": access, "refreshToken": newRefresh})
	})
}
