package integration

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gabe-dev-svc/volley/internal/api"
	"github.com/gabe-dev-svc/volley/internal/repository"
	"github.com/gabe-dev-svc/volley/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testServer     *http.Server
	testBaseURL    string
	testDBPool     *pgxpool.Pool
	testContainer  *postgres.PostgresContainer
	testCtx        context.Context
)

func TestMain(m *testing.M) {
	var err error
	testCtx = context.Background()

	// Set up logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Start PostgreSQL container with PostGIS
	testContainer, err = postgres.Run(testCtx,
		"postgis/postgis:16-3.4",
		postgres.WithDatabase("volley_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start PostgreSQL container")
	}

	// Get connection string
	connStr, err := testContainer.ConnectionString(testCtx, "sslmode=disable")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get connection string")
	}

	// Set DATABASE_URL for migrations
	os.Setenv("DATABASE_URL", connStr)

	// Connect to database
	testDBPool, err = pgxpool.New(testCtx, connStr)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to test database")
	}

	// Enable PostGIS extension
	_, err = testDBPool.Exec(testCtx, "CREATE EXTENSION IF NOT EXISTS postgis")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to enable PostGIS extension")
	}

	// Run migrations
	if err := runMigrations(testCtx, testDBPool); err != nil {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}

	// Start test API server
	queries := repository.New(testDBPool)
	gamesService := service.NewGamesService(queries, testDBPool)
	userService := service.NewUserService(queries)
	handler := api.NewHandler(gamesService, userService)

	// Set up router with middleware
	router := gin.New()
	router.Use(api.RequestIDMiddleware())
	router.Use(gin.Recovery())
	handler.RegisterRoutes(router)

	// Start server on fixed port for simplicity
	testBaseURL = "http://localhost:8081"
	testServer = &http.Server{
		Addr:    ":8081",
		Handler: router,
	}

	// Start server in background
	go func() {
		if err := testServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("Test server failed")
		}
	}()

	// Wait for server to start
	time.Sleep(500 * time.Millisecond)

	// Run tests
	code := m.Run()

	// Cleanup
	if err := testServer.Shutdown(testCtx); err != nil {
		log.Error().Err(err).Msg("Failed to shutdown test server")
	}
	testDBPool.Close()
	if err := testContainer.Terminate(testCtx); err != nil {
		log.Error().Err(err).Msg("Failed to terminate container")
	}

	os.Exit(code)
}

// runMigrations applies the database schema
func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	// Read schema file
	schemaPath := "../../internal/repository/schema.sql"
	schemaSQL, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("read schema file: %w", err)
	}

	// Execute schema
	_, err = pool.Exec(ctx, string(schemaSQL))
	if err != nil {
		return fmt.Errorf("execute schema: %w", err)
	}

	log.Info().Msg("Database migrations completed successfully")
	return nil
}
