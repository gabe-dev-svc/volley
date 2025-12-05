package api

import (
	"context"
	"os"
	"time"

	"github.com/gabe-dev-svc/volley/internal/database"
	"github.com/gabe-dev-svc/volley/internal/repository"
	"github.com/gabe-dev-svc/volley/internal/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Server struct {
	router  *gin.Engine
	handler *Handler
}

func NewServer() *Server {
	// Configure zerolog based on environment
	configureLogger()

	// Initialize database pool
	ctx := context.Background()
	pool, err := database.NewPool(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create database pool")
	}

	// Create repository queries
	queries := repository.New(pool)

	// Initialize services with repository
	gamesService := service.NewGamesService(queries, pool)
	userService := service.NewUserService(queries)

	// Load Google Places API key from environment
	googlePlacesKey := os.Getenv("GOOGLE_PLACES_API_KEY")
	if googlePlacesKey == "" {
		log.Warn().Msg("GOOGLE_PLACES_API_KEY not set - location autocomplete will not work")
	}

	// Set up router
	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(CustomGinLogger())
	router.Use(gin.Recovery())

	// Configure CORS to allow all localhost origins for development
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = append(config.AllowHeaders, "X-Client-Type", "Authorization")
	router.Use(cors.New(config))

	handler := NewHandler(gamesService, userService, googlePlacesKey)
	handler.RegisterRoutes(router)

	log.Info().Msg("Server initialized successfully")

	return &Server{
		router:  router,
		handler: handler,
	}
}

func CustomGinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()
		requestID, _ := c.Get("requestID")

		if raw != "" {
			path = path + "?" + raw
		}

		logger := log.With().
			Str("requestId", requestID.(string)).
			Str("method", method).
			Str("path", path).
			Str("clientIp", clientIP).
			Str("userAgent", userAgent).
			Logger()

		logger.Info().Msg("Begin HTTP request")

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		event := logger.Info()
		if status >= 500 {
			event = logger.Error()
		} else if status >= 400 {
			event = logger.Warn()
		}

		event.
			Int("status", status).
			Int64("latencyMs", latency.Milliseconds()).
			Int("bodySize", c.Writer.Size()).
			Msg("End HTTP request")
	}
}

// configureLogger sets up zerolog with appropriate settings for the environment
func configureLogger() {
	// Set global log level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if os.Getenv("GIN_MODE") != "release" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		// Use console writer for development
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// Add service name to all logs
	log.Logger = log.With().Str("service", "volley-api").Logger()
}

func (s *Server) Run(port string) error {
	return s.router.Run(":" + port)
}
