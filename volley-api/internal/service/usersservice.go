package service

import (
	"context"
	"fmt"

	"github.com/gabe-dev-svc/volley/ifaces"
	"github.com/gabe-dev-svc/volley/internal/errors"
	"github.com/gabe-dev-svc/volley/internal/models"
	"github.com/gabe-dev-svc/volley/internal/repository"
	"github.com/gabe-dev-svc/volley/internal/util"
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
