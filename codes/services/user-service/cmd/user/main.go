package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"

	"github.com/venue-master/platform/internal/server"
	"github.com/venue-master/platform/lib/config"
	"github.com/venue-master/platform/services/user-service/internal/store"
)

type authRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func main() {
	srv, err := server.New("user-service")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	repo, err := initStore(ctx, srv.Config.Database, srv.Logger)
	if err != nil {
		panic(err)
	}
	defer repo.Close()

	if err := seedDefaultUser(ctx, repo, srv.Logger); err != nil {
		panic(err)
	}

	registerRoutes(srv.Engine, repo)

	if err := srv.Run(); err != nil {
		panic(err)
	}
}

func seedDefaultUser(ctx context.Context, repo *store.Store, logger zerolog.Logger) error {
	email := os.Getenv("DEFAULT_MEMBER_EMAIL")
	password := os.Getenv("DEFAULT_MEMBER_PASSWORD")
	if email == "" || password == "" {
		logger.Warn().Msg("DEFAULT_MEMBER_EMAIL/PASSWORD not set; skipping seed")
		return nil
	}
	if _, err := repo.SeedDefaultUser(ctx, email, password); err != nil {
		return err
	}
	logger.Info().Str("email", email).Msg("ensured default member exists")
	return nil
}

func registerRoutes(router *gin.Engine, repo *store.Store) {
	group := router.Group("/v1/users")

	group.GET("/me", func(ctx *gin.Context) {
		userID := ctx.GetHeader("X-User-ID")
		if userID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "X-User-ID header required"})
			return
		}
		handleGetUser(ctx, repo, userID)
	})

	group.GET("/:id", func(ctx *gin.Context) {
		handleGetUser(ctx, repo, ctx.Param("id"))
	})

	group.GET("/:id/memberships", func(ctx *gin.Context) {
		user, err := fetchUser(ctx, repo, ctx.Param("id"))
		if err != nil {
			handleStoreError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, []gin.H{
			{
				"id":         "membership-" + user.ID.String(),
				"type":       "MONTHLY_PREMIUM",
				"status":     "ACTIVE",
				"startDate":  user.CreatedAt.Format(time.RFC3339),
				"expiryDate": user.CreatedAt.AddDate(0, 1, 0).Format(time.RFC3339),
				"autoRenew":  true,
			},
		})
	})

	group.POST("/authenticate", func(ctx *gin.Context) {
		var req authRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		timeoutCtx, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
		defer cancel()

		user, err := repo.GetUserByEmail(timeoutCtx, req.Email)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		if err := store.ComparePassword(user.PasswordHash, req.Password); err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		ctx.JSON(http.StatusOK, userResponse(user))
	})

	group.POST("/register", func(ctx *gin.Context) {
		var req struct {
			Email     string `json:"email" binding:"required,email"`
			Password  string `json:"password" binding:"required,min=8"`
			FirstName string `json:"firstName" binding:"required"`
			LastName  string `json:"lastName" binding:"required"`
			Phone     string `json:"phone"`
		}
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		timeoutCtx, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
		defer cancel()

		// Check if email already exists
		existingUser, err := repo.GetUserByEmail(timeoutCtx, req.Email)
		if err == nil && existingUser != nil {
			ctx.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			return
		}
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}

		// Hash password
		hash, err := store.HashPassword(req.Password)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "password hashing failed"})
			return
		}

		// Create new user
		user := &store.User{
			ID:           uuid.New(),
			Email:        req.Email,
			FirstName:    req.FirstName,
			LastName:     req.LastName,
			Roles:        []string{"MEMBER"},
			PasswordHash: hash,
		}

		if err := repo.UpsertUser(timeoutCtx, user); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}

		ctx.JSON(http.StatusCreated, userResponse(user))
	})
}

func handleGetUser(ctx *gin.Context, repo *store.Store, idParam string) {
	user, err := fetchUser(ctx, repo, idParam)
	if err != nil {
		handleStoreError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, userResponse(user))
}

func fetchUser(ctx *gin.Context, repo *store.Store, idParam string) (*store.User, error) {
	id, err := uuid.Parse(idParam)
	if err != nil {
		return nil, err
	}
	timeoutCtx, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()
	return repo.GetUserByID(timeoutCtx, id)
}

func handleStoreError(ctx *gin.Context, err error) {
	if errors.Is(err, pgx.ErrNoRows) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

func userResponse(user *store.User) gin.H {
	return gin.H{
		"id":        user.ID.String(),
		"email":     user.Email,
		"firstName": user.FirstName,
		"lastName":  user.LastName,
		"roles":     user.Roles,
		"createdAt": user.CreatedAt.Format(time.RFC3339),
		"updatedAt": user.UpdatedAt.Format(time.RFC3339),
	}
}

const (
	dbReadyTimeout = 60 * time.Second
	dbRetryDelay   = 3 * time.Second
)

func initStore(ctx context.Context, cfg config.DatabaseConfig, logger zerolog.Logger) (*store.Store, error) {
	deadline := time.Now().Add(dbReadyTimeout)
	var lastErr error

	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		attemptCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		repo, err := store.New(attemptCtx, cfg)
		cancel()
		if err == nil {
			if err = repo.Ping(ctx); err == nil {
				if err = repo.RunMigrations(ctx); err == nil {
					logger.Info().Msg("connected to Postgres")
					return repo, nil
				}
				err = fmt.Errorf("migration error: %w", err)
			} else {
				err = fmt.Errorf("ping error: %w", err)
			}
			repo.Close()
		}

		lastErr = err
		if time.Now().After(deadline) {
			break
		}

		logger.Warn().Err(err).Msg("Postgres not ready, retrying…")
		time.Sleep(dbRetryDelay)
	}

	return nil, lastErr
}
