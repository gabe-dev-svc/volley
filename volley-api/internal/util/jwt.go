package util

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// JWTClaims represents the claims stored in the JWT
type JWTClaims struct {
	UserID    string `json:"userId"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	jwt.RegisteredClaims
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SecretKey       string
	ExpirationHours int
}

// DefaultJWTConfig returns default JWT configuration
func DefaultJWTConfig() *JWTConfig {
	return &JWTConfig{
		SecretKey:       "your-secret-key-change-this-in-production", // TODO: Load from env
		ExpirationHours: 24 * 7,                                      // 7 days
	}
}

// GenerateToken creates a new JWT token for a user
func GenerateToken(userID, email, firstName, lastName string, config *JWTConfig) (string, error) {
	if config == nil {
		config = DefaultJWTConfig()
	}

	claims := JWTClaims{
		UserID:    userID,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(config.ExpirationHours))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "volley-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.SecretKey))
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(c context.Context, tokenString string) (*JWTClaims, error) {
	logger := log.With().Caller().Logger()

	config := DefaultJWTConfig()

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(config.SecretKey), nil
	})

	if err != nil {
		logger.Error().Str("token", tokenString).Err(err).Msg("ParseWithClaims failed")
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
