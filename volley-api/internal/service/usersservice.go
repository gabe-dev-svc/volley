package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gabe-dev-svc/volley/ifaces"
	"github.com/gabe-dev-svc/volley/internal/errors"
	"github.com/gabe-dev-svc/volley/internal/models"
	"github.com/gabe-dev-svc/volley/internal/repository"
	"github.com/gabe-dev-svc/volley/internal/util"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
)

type UserService struct {
	queries ifaces.Querier
}

func NewUserService(queries ifaces.Querier) *UserService {
	return &UserService{
		queries: queries,
	}
}

type CreateUserRequest struct {
	FirstName string
	LastName  string
	Password  string
	Email     string
}

func (u *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (*models.User, error) {
	logger := log.Ctx(ctx)

	// Check if user with this email already exists
	existingUser, err := u.queries.GetUserByEmail(ctx, req.Email)
	if err == nil && existingUser.ID.Valid {
		logger.Warn().Str("email", req.Email).Msg("User with this email already exists")
		return nil, fmt.Errorf("user with email %s: %w", req.Email, errors.ErrAlreadyExists)
	}

	// Create user object
	user := &models.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
	}

	// Hash the password using Argon2
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		logger.Error().Err(err).Msg("HashPassword failed")
		return nil, fmt.Errorf("HashPassword failed: %w", err)
	}

	newUser, err := u.queries.CreateUser(ctx, repository.CreateUserParams{
		Email:        req.Email,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		PasswordHash: hashedPassword,
	})
	if err != nil {
		logger.Error().Err(err).Msg("CreateUser failed")
		return nil, fmt.Errorf("CreateUser failed: %w", err)
	}
	user.ID = newUser.ID.String()
	user.CreatedAt = newUser.CreatedAt.Time

	// Print for debugging (remove in production)
	fmt.Printf("User registered: %s %s (%s)\n", req.FirstName, req.LastName, req.Email)
	fmt.Printf("Password hash: %s\n", hashedPassword)

	return user, nil
}

func (u *UserService) Login(ctx context.Context, email string, password string) (*models.User, error) {
	logger := log.Ctx(ctx)

	// Get user by email
	dbUser, err := u.queries.GetUserByEmail(ctx, email)
	if err != nil {
		logger.Warn().Str("email", email).Msg("User not found")
		return nil, fmt.Errorf("invalid email or password")
	}

	// Verify password
	valid, err := util.VerifyPassword(password, dbUser.PasswordHash)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to verify password")
		return nil, fmt.Errorf("invalid email or password")
	}
	if !valid {
		logger.Warn().Str("email", email).Msg("Invalid password")
		return nil, fmt.Errorf("invalid email or password")
	}

	user := &models.User{
		ID:        dbUser.ID.String(),
		Email:     dbUser.Email,
		FirstName: dbUser.FirstName,
		LastName:  dbUser.LastName,
		CreatedAt: dbUser.CreatedAt.Time,
	}

	logger.Info().Str("email", email).Msg("User logged in successfully")
	return user, nil
}

// CreateRefreshTokenForUser creates a new refresh token for a user
// Returns the token (to send to client) and stores the hash in the database
func (u *UserService) CreateRefreshTokenForUser(ctx context.Context, userID string, deviceInfo string) (string, error) {
	logger := log.Ctx(ctx)

	// Parse UUID
	var userUUID pgtype.UUID
	if err := userUUID.Scan(userID); err != nil {
		return "", fmt.Errorf("invalid user ID: %w", err)
	}

	// Generate refresh token
	token, tokenHash, err := util.GenerateRefreshToken()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to generate refresh token")
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store token hash in database (expires in 30 days)
	expiresAt := pgtype.Timestamptz{
		Time:  time.Now().Add(30 * 24 * time.Hour),
		Valid: true,
	}

	_, err = u.queries.CreateRefreshToken(ctx, repository.CreateRefreshTokenParams{
		UserID: userUUID,
		TokenHash: tokenHash,
		DeviceInfo: pgtype.Text{
			String: deviceInfo,
			Valid:  deviceInfo != "",
		},
		ExpiresAt: expiresAt,
	})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to store refresh token")
		return "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	logger.Info().Str("userID", userID).Msg("Refresh token created")
	return token, nil
}

// ValidateRefreshToken validates a refresh token and returns the associated user
func (u *UserService) ValidateRefreshToken(ctx context.Context, token string) (*models.User, error) {
	logger := log.Ctx(ctx)

	// Hash the token to look it up
	tokenHash := util.HashRefreshToken(token)

	// Get token from database (will only return if valid and not expired)
	dbToken, err := u.queries.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		logger.Warn().Msg("Invalid or expired refresh token")
		return nil, fmt.Errorf("invalid or expired refresh token")
	}

	// Get the user
	dbUser, err := u.queries.GetUserByID(ctx, dbToken.UserID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get user for refresh token")
		return nil, fmt.Errorf("failed to get user")
	}

	user := &models.User{
		ID:        dbUser.ID.String(),
		Email:     dbUser.Email,
		FirstName: dbUser.FirstName,
		LastName:  dbUser.LastName,
		CreatedAt: dbUser.CreatedAt.Time,
	}

	return user, nil
}

// RevokeRefreshToken revokes a specific refresh token
func (u *UserService) RevokeRefreshToken(ctx context.Context, token string) error {
	tokenHash := util.HashRefreshToken(token)
	return u.queries.RevokeRefreshToken(ctx, tokenHash)
}

// RevokeAllUserRefreshTokens revokes all refresh tokens for a user (useful for logout from all devices)
func (u *UserService) RevokeAllUserRefreshTokens(ctx context.Context, userID string) error {
	var userUUID pgtype.UUID
	if err := userUUID.Scan(userID); err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}
	return u.queries.RevokeAllUserRefreshTokens(ctx, userUUID)
}
