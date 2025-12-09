package util

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

const (
	// RefreshTokenLength is the number of random bytes in a refresh token (256 bits = 32 bytes)
	RefreshTokenLength = 32
)

// GenerateRefreshToken generates a cryptographically secure random refresh token
// Returns the token (to give to client) and its SHA-256 hash (to store in database)
func GenerateRefreshToken() (token string, tokenHash string, err error) {
	// Generate 32 random bytes
	bytes := make([]byte, RefreshTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode as base64 for transmission
	token = base64.URLEncoding.EncodeToString(bytes)

	// Hash the token for storage (we never store the raw token)
	hash := sha256.Sum256([]byte(token))
	tokenHash = base64.URLEncoding.EncodeToString(hash[:])

	return token, tokenHash, nil
}

// HashRefreshToken hashes a refresh token using SHA-256
// This is used to hash tokens received from clients before database lookup
func HashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(hash[:])
}
