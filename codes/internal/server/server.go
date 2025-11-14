package server

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/venue-master/platform/lib/config"
	"github.com/venue-master/platform/lib/logutil"
)

// Server bundles the gin engine with shared config + logger.
type Server struct {
	Engine *gin.Engine
	Config *config.Config
	Logger zerolog.Logger
}

// New constructs a Server with common middleware.
func New(serviceName string) (*Server, error) {
	cfg, err := config.Load(serviceName)
	if err != nil {
		return nil, err
	}

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	logger := logutil.New(serviceName, cfg.AppEnv)

	// CORS middleware - allow all origins, no credentials
	corsConfig := cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type", "X-User-ID", "X-User-Roles"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}

	engine.Use(gin.Recovery(), requestLogger(logger), cors.New(corsConfig))
	engine.GET("/healthz", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok", "service": serviceName})
	})

	return &Server{Engine: engine, Config: cfg, Logger: logger}, nil
}

// Run boots the HTTP server.
func (s *Server) Run() error {
	s.Logger.Info().Str("addr", s.Config.HTTPAddr()).Msg("server starting")
	return s.Engine.Run(s.Config.HTTPAddr())
}

func requestLogger(logger zerolog.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		ctx.Next()
		logger.Info().
			Str("method", ctx.Request.Method).
			Str("path", ctx.Request.URL.Path).
			Int("status", ctx.Writer.Status()).
			Dur("duration", time.Since(start)).
			Msg("http")
	}
}
