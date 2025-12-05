package api

import (
	"errors"
	"net/http"
	"strings"

	volleyerrors "github.com/gabe-dev-svc/volley/internal/errors"
	"github.com/gabe-dev-svc/volley/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// LoggerFromContext returns a zerolog logger with request context fields
func LoggerFromContext(c *gin.Context) zerolog.Logger {
	requestID := c.GetString("requestID")
	return log.With().
		Caller().
		Str("requestId", requestID).
		Str("requestURI", c.Request.RequestURI).
		Logger()
}

// RequestIDMiddleware generates a unique request ID for each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Set("requestID", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

type tokenFromRequest struct {
	token       string
	tokenSource string
}

func getTokenFromRequest(c *gin.Context) (tokenFromRequest, error) {
	logger := LoggerFromContext(c)

	var token string
	var tokenSource string

	// Use same logic as registration/login for detecting client type
	if isMobileClient(c) {
		// For mobile clients: get token from Authorization header
		tokenSource = "header"
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			// Expected format: "Bearer <token>"
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				token = parts[1]
			}
		}
	} else {
		// For web clients: get token from cookie
		tokenSource = "cookie"
		var err error
		token, err = c.Cookie("auth_token")

		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				return tokenFromRequest{}, volleyerrors.ErrMissingAuthToken
			}
			logger.Error().Err(err).Msg("failed getting auth_token from cookies")
			return tokenFromRequest{}, volleyerrors.ErrInvalidAuthToken
		}
	}

	if token == "" {
		return tokenFromRequest{}, volleyerrors.ErrMissingAuthToken
	}

	return tokenFromRequest{token: token, tokenSource: tokenSource}, nil
}

func authenticateUser(c *gin.Context) error {
	logger := LoggerFromContext(c)

	tokenFromRequest, err := getTokenFromRequest(c)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get token from request")
		return err
	}

	token, tokenSource := tokenFromRequest.token, tokenFromRequest.tokenSource

	if token == "" {
		logger.Error().Msg("authentication token is empty")
		return volleyerrors.ErrMissingAuthToken
	}

	// Validate the token
	claims, err := util.ValidateToken(c, token)
	if err != nil {
		logger.Warn().
			Err(err).
			Str("reason", "invalid_token").
			Str("tokenSource", tokenSource).
			Msg("Authentication failed")
		return volleyerrors.ErrInvalidAuthToken
	}

	// Validate claims have required fields
	if claims.UserID == "" {
		logger.Error().Msg("Token missing userID claim")
		return volleyerrors.ErrInvalidAuthToken
	}

	// Set user information in context
	c.Set("userID", claims.UserID)
	c.Set("email", claims.Email)
	c.Set("firstName", claims.FirstName)
	c.Set("lastName", claims.LastName)

	logger.Debug().
		Str("userID", claims.UserID).
		Str("email", claims.Email).
		Msg("User authenticated successfully")

	return nil
}

func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := LoggerFromContext(c)
		if err := authenticateUser(c); err != nil {
			if !errors.Is(err, volleyerrors.ErrMissingAuthToken) {
				logger.Error().Err(err).Msg("failed to authenticate user")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication token provided"})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

// AuthMiddleware validates JWT token and sets user information in context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := LoggerFromContext(c)
		if err := authenticateUser(c); err != nil {
			errorMessage := "Missing authentication token"
			if errors.Is(err, volleyerrors.ErrInvalidAuthToken) {
				errorMessage = "Invalid authentication token provided"
			}
			logger.Error().Err(err).Str("reason", errorMessage).Msg("Authentication failed")
			c.JSON(http.StatusUnauthorized, gin.H{"error": errorMessage})
			c.Abort()
			return
		}

		c.Next()
	}
}
