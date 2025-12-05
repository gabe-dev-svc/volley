package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gabe-dev-svc/volley/ifaces"
	apperrors "github.com/gabe-dev-svc/volley/internal/errors"
	"github.com/gabe-dev-svc/volley/internal/models"
	"github.com/gabe-dev-svc/volley/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type InvalidArgumentError struct {
	ArgumentName string
	Message      string
}

func (i *InvalidArgumentError) Error() string {
	return fmt.Sprintf("Invalid argument provided: %s - %s", i.ArgumentName, i.Message)
}

func NewInvalidArgumentError(argumentName string, message string) InvalidArgumentError {
	return InvalidArgumentError{
		ArgumentName: argumentName,
		Message:      message,
	}
}

var (

	// Inactive participant states
	InactiveParticipantStates = map[string]bool{
		string(models.ParticipantStatusDeclined): true,
		string(models.ParticipantStatusDropped):  true,
		string(models.ParticipantStatusRemoved):  true,
	}

	// Custom error types
	ErrInvalidLatitude    = NewInvalidArgumentError("latitude", "latitude must be between -90 and 90")
	ErrInvalidLongitude   = NewInvalidArgumentError("longitude", "longitude must be between -180 and 180")
	ErrInvalidRadius      = NewInvalidArgumentError("radius", "radius must be non-negative")
	ErrMissingStartDate   = NewInvalidArgumentError("start_date", "start_date is required")
	ErrTooLate            = errors.New("too late to drop from game")
	ErrGameFinished       = errors.New("game has already finished")
	ErrNotParticipant     = errors.New("user is not a participant of this game")
	ErrNotOwner           = errors.New("only the game owner can cancel the game")
	ErrAlreadyCancelled   = errors.New("game is already cancelled")
	ErrGameAlreadyStarted = errors.New("cannot cancel a game that has already started")
)

type GamesService struct {
	queries ifaces.Querier
	pool    *pgxpool.Pool
}

func NewGamesService(queries ifaces.Querier, pool *pgxpool.Pool) *GamesService {
	return &GamesService{
		queries: queries,
		pool:    pool,
	}
}

type TimeFilter string

const (
	TimeFilterUpcoming TimeFilter = "upcoming"
	TimeFilterPast     TimeFilter = "past"
	TimeFilterAll      TimeFilter = "all"
)

type ListGamesFilters struct {
	Categories []string   // Sport categories (required: soccer, basketball, volleyball, etc.)
	Latitude   float64    // Latitude coordinate for location-based search (required)
	Longitude  float64    // Longitude coordinate for location-based search (required)
	Radius     float64    // Search radius in meters (default: 16093.4 meters = 10 miles)
	TimeFilter TimeFilter // Filter by time: upcoming, past, all (default: upcoming)
	Status     *string    // Filter by game status (open, full, closed, etc.)
	Limit      int        // Number of results to return (default 20, max 100)
	Offset     int        // Number of results to skip (default 0)
}

// ListGames retrieves a list of games based on filters
func (s *GamesService) ListGames(ctx context.Context, filters ListGamesFilters, userID *string) ([]models.GameSummary, error) {
	// Validate required fields
	if len(filters.Categories) == 0 {
		return nil, &InvalidArgumentError{
			ArgumentName: "categories",
			Message:      "at least one sport category is required",
		}
	}
	if filters.Latitude < -90 || filters.Latitude > 90 {
		return nil, &ErrInvalidLatitude
	}
	if filters.Longitude < -180 || filters.Longitude > 180 {
		return nil, &ErrInvalidLongitude
	}
	if filters.Radius < 0 {
		return nil, &ErrInvalidRadius
	}

	// Set defaults
	if filters.Radius == 0 {
		filters.Radius = 16093.4 // 10 miles in meters
	}
	if filters.TimeFilter == "" {
		filters.TimeFilter = TimeFilterUpcoming
	}
	if filters.Limit == 0 {
		filters.Limit = 20
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}

	// Determine time range based on filter
	now := time.Now()
	var startTime time.Time
	var endTime pgtype.Timestamptz

	switch filters.TimeFilter {
	case TimeFilterUpcoming:
		startTime = now // Games starting now or in the future
		endTime = pgtype.Timestamptz{Valid: false}
	case TimeFilterPast:
		startTime = time.Time{} // Beginning of time
		endTime = pgtype.Timestamptz{Time: now, Valid: true}
	case TimeFilterAll:
		startTime = time.Time{} // All games
		endTime = pgtype.Timestamptz{Valid: false}
	}

	// Query games with category filter
	statusText := pgtype.Text{Valid: false}
	if filters.Status != nil {
		statusText = pgtype.Text{String: *filters.Status, Valid: true}
	}

	// Handle optional user ID for participation status
	var userUUID pgtype.UUID
	if userID != nil {
		if err := userUUID.Scan(*userID); err != nil {
			return nil, &InvalidArgumentError{
				ArgumentName: "user_id",
				Message:      "invalid user ID format",
			}
		}
	} else {
		userUUID = pgtype.UUID{Valid: false}
	}

	games, err := s.queries.ListGamesInRadius(ctx, repository.ListGamesInRadiusParams{
		Longitude:  filters.Longitude,
		Latitude:   filters.Latitude,
		Radius:     filters.Radius,
		StartTime:  pgtype.Timestamptz{Time: startTime, Valid: true},
		EndTime:    endTime,
		Status:     statusText,
		Categories: filters.Categories,
		UserID:     userUUID,
		Limit:      int32(filters.Limit),
		Offset:     int32(filters.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list games: %w", err)
	}

	// Convert repository games to model game summaries
	var allGames []models.GameSummary
	for _, game := range games {
		allGames = append(allGames, convertGameRowToSummary(game))
	}

	return allGames, nil
}

// convertGameRowToSummary converts a repository game row to a models.GameSummary
func convertGameRowToSummary(g repository.ListGamesInRadiusRow) models.GameSummary {
	lat := g.Latitude.(float64)
	lng := g.Longitude.(float64)

	// Convert user participation status if present
	log.Debug().Interface("gameRow", g).Send()
	var userParticipationStatus *models.ParticipantStatus
	if g.UserParticipationStatus.Valid && g.UserParticipationStatus.String != "" {
		status := models.ParticipantStatus(g.UserParticipationStatus.String)
		userParticipationStatus = &status
	}

	return models.GameSummary{
		ID:          g.ID.String(),
		Category:    models.GameCategory(g.Category),
		Title:       pgTextToStringPtr(g.Title),
		Description: pgTextToStringPtr(g.Description),
		Location: models.Location{
			Name:      g.LocationName,
			Address:   pgTextToStringPtr(g.LocationAddress),
			Latitude:  &lat,
			Longitude: &lng,
			Notes:     pgTextToStringPtr(g.LocationNotes),
		},
		StartTime:           g.StartTime.Time,
		DurationMinutes:     int(g.DurationMinutes),
		MaxParticipants:     int(g.MaxParticipants),
		CurrentParticipants: int(g.CurrentParticipants),
		Pricing: models.Pricing{
			Type:        models.PricingType(g.PricingType),
			AmountCents: int(g.PricingAmountCents),
			Currency:    g.PricingCurrency,
		},
		SignupDeadline:          g.SignupDeadline.Time,
		SkillLevel:              models.SkillLevel(g.SkillLevel),
		Status:                  models.GameStatus(g.Status),
		UserParticipationStatus: userParticipationStatus,
	}
}

func pgTimestamptzToTimePtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

// CreateGame creates a new game
func (s *GamesService) CreateGame(ctx context.Context, userID string, request models.CreateGameRequest) (*models.Game, error) {
	var ownerID pgtype.UUID
	if err := ownerID.Scan(userID); err != nil {
		return nil, &InvalidArgumentError{
			ArgumentName: "user_id",
			Message:      "invalid user ID format",
		}
	}

	// Set defaults
	signupDeadline := request.StartTime
	if request.SignupDeadline != nil {
		signupDeadline = *request.SignupDeadline
	}

	skillLevel := models.SkillLevelAll
	if request.SkillLevel != nil {
		skillLevel = *request.SkillLevel
	}

	// Validate location coordinates are provided
	if request.Location.Latitude == nil || request.Location.Longitude == nil {
		return nil, &InvalidArgumentError{
			ArgumentName: "location",
			Message:      "location latitude and longitude are required",
		}
	}

	createGameRequest := repository.CreateGameParams{
		OwnerID:  ownerID,
		Category: string(request.Category),
		Title: pgtype.Text{
			String: stringPtrToString(request.Title),
			Valid:  request.Title != nil,
		},
		Description: pgtype.Text{
			String: stringPtrToString(request.Description),
			Valid:  request.Description != nil,
		},
		LocationName: request.Location.Name,
		LocationAddress: pgtype.Text{
			String: stringPtrToString(request.Location.Address),
			Valid:  request.Location.Address != nil,
		},
		Longitude: *request.Location.Longitude,
		Latitude:  *request.Location.Latitude,
		LocationNotes: pgtype.Text{
			String: stringPtrToString(request.Location.Notes),
			Valid:  request.Location.Notes != nil,
		},
		StartTime: pgtype.Timestamptz{
			Time:  request.StartTime,
			Valid: true,
		},
		DurationMinutes:    int32(request.DurationMinutes),
		MaxParticipants:    int32(request.MaxParticipants),
		PricingType:        string(request.Pricing.Type),
		PricingAmountCents: int32(request.Pricing.AmountCents),
		PricingCurrency:    request.Pricing.Currency,
		SignupDeadline: pgtype.Timestamptz{
			Time:  signupDeadline,
			Valid: true,
		},
		DropDeadline: pgtype.Timestamptz{
			Time: func() time.Time {
				if request.DropDeadline != nil {
					return *request.DropDeadline
				}
				return time.Time{}
			}(),
			Valid: request.DropDeadline != nil,
		},
		SkillLevel: string(skillLevel),
		Notes: pgtype.Text{
			String: stringPtrToString(request.Notes),
			Valid:  request.Notes != nil,
		},
		Status: string(models.GameStatusOpen),
	}

	game, err := s.queries.CreateGame(ctx, createGameRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	return convertCreateGameRowToModel(game), nil
}

// convertCreateGameRowToModel converts a repository.CreateGameRow to a models.Game
func convertCreateGameRowToModel(game repository.CreateGameRow) *models.Game {
	var lat, lng *float64
	if v, ok := game.Latitude.(float64); ok {
		lat = &v
	}
	if v, ok := game.Longitude.(float64); ok {
		lng = &v
	}

	return &models.Game{
		ID:          uuid.UUID(game.ID.Bytes).String(),
		Category:    models.GameCategory(game.Category),
		Title:       pgTextToStringPtr(game.Title),
		Description: pgTextToStringPtr(game.Description),
		Location: models.Location{
			Name:      game.LocationName,
			Address:   pgTextToStringPtr(game.LocationAddress),
			Latitude:  lat,
			Longitude: lng,
			Notes:     pgTextToStringPtr(game.LocationNotes),
		},
		StartTime:       game.StartTime.Time.UTC(),
		DurationMinutes: int(game.DurationMinutes),
		MaxParticipants: int(game.MaxParticipants),
		Pricing: models.Pricing{
			Type:        models.PricingType(game.PricingType),
			AmountCents: int(game.PricingAmountCents),
			Currency:    game.PricingCurrency,
		},
		SignupDeadline: game.SignupDeadline.Time.UTC(),
		SkillLevel:     models.SkillLevel(game.SkillLevel),
		Notes:          pgTextToStringPtr(game.Notes),
		Status:         models.GameStatus(game.Status),
		CreatedAt:      game.CreatedAt.Time.UTC(),
		UpdatedAt:      game.UpdatedAt.Time.UTC(),
	}
}

// convertGetGameRowToModel converts a repository.GetGameRow to a models.Game with participants
func convertGetGameRowToModel(game repository.GetGameRow, owner *repository.User, confirmedParticipants []models.Participant, waitlist []models.Participant) *models.Game {
	var lat, lng *float64
	if v, ok := game.Latitude.(float64); ok {
		lat = &v
	}
	if v, ok := game.Longitude.(float64); ok {
		lng = &v
	}

	var ownerModel *models.User
	if owner != nil {
		ownerModel = &models.User{
			ID:        uuid.UUID(owner.ID.Bytes).String(),
			Email:     owner.Email,
			FirstName: owner.FirstName,
			LastName:  owner.LastName,
			CreatedAt: owner.CreatedAt.Time.UTC(),
		}
	}

	return &models.Game{
		ID:          uuid.UUID(game.ID.Bytes).String(),
		Owner:       ownerModel,
		Category:    models.GameCategory(game.Category),
		Title:       pgTextToStringPtr(game.Title),
		Description: pgTextToStringPtr(game.Description),
		Location: models.Location{
			Name:      game.LocationName,
			Address:   pgTextToStringPtr(game.LocationAddress),
			Latitude:  lat,
			Longitude: lng,
			Notes:     pgTextToStringPtr(game.LocationNotes),
		},
		StartTime:             game.StartTime.Time.UTC(),
		DurationMinutes:       int(game.DurationMinutes),
		MaxParticipants:       int(game.MaxParticipants),
		ConfirmedParticipants: confirmedParticipants,
		Waitlist:              waitlist,
		Pricing: models.Pricing{
			Type:        models.PricingType(game.PricingType),
			AmountCents: int(game.PricingAmountCents),
			Currency:    game.PricingCurrency,
		},
		SignupDeadline: game.SignupDeadline.Time.UTC(),
		DropDeadline:   pgTimestamptzToTimePtr(game.DropDeadline),
		SkillLevel:     models.SkillLevel(game.SkillLevel),
		Notes:          pgTextToStringPtr(game.Notes),
		Status:         models.GameStatus(game.Status),
		CancelledAt:    pgTimestamptzToTimePtr(game.CancelledAt),
		CreatedAt:      game.CreatedAt.Time.UTC(),
		UpdatedAt:      game.UpdatedAt.Time.UTC(),
	}
}

// Helper function to convert pgtype.Text to *string
func pgTextToStringPtr(text pgtype.Text) *string {
	if !text.Valid {
		return nil
	}
	return &text.String
}

// Helper function to safely dereference string pointers
func stringPtrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// GetGame retrieves a single game by ID with full participant details
func (s *GamesService) GetGame(ctx context.Context, gameID string) (*models.Game, error) {
	// Validate game UUID
	var gameUUID pgtype.UUID
	if err := gameUUID.Scan(gameID); err != nil {
		return nil, &InvalidArgumentError{
			ArgumentName: "game_id",
			Message:      "invalid game ID format",
		}
	}

	// Get game details
	gameRow, err := s.queries.GetGame(ctx, gameUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	// Get owner details (required)
	owner, err := s.queries.GetUserByID(ctx, gameRow.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game owner: %w", err)
	}

	// Get all participants ordered by joined_at
	allParticipants, err := s.queries.ListParticipantsByGame(ctx, gameUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to list participants: %w", err)
	}

	// Split confirmed participants into roster and waitlist based on MaxParticipants
	confirmedParticipants := []models.Participant{}
	waitlist := []models.Participant{}
	confirmedCount := 0

	for _, p := range allParticipants {
		// Skip non-confirmed participants
		if p.Status != string(models.ParticipantStatusConfirmed) {
			continue
		}

		if confirmedCount < int(gameRow.MaxParticipants) {
			// First N confirmed participants are on the roster
			participant := convertParticipantRowToModelWithStatus(p, models.ParticipantStatusConfirmed, nil)
			confirmedParticipants = append(confirmedParticipants, *participant)
		} else {
			// Beyond N confirmed participants are waitlisted
			position := confirmedCount - int(gameRow.MaxParticipants) + 1
			participant := convertParticipantRowToModelWithStatus(p, models.ParticipantStatusWaitlist, &position)
			waitlist = append(waitlist, *participant)
		}
		confirmedCount++
	}

	// Convert to Game model
	game := convertGetGameRowToModel(gameRow, &owner, confirmedParticipants, waitlist)
	return game, nil
}

// UpdateGame updates an existing game
func (s *GamesService) UpdateGame(ctx context.Context, gameID string, userID string, request interface{}) (interface{}, error) {
	// TODO: Implement game update logic
	return nil, nil
}

// DeleteGame deletes/cancels a game
func (s *GamesService) DeleteGame(ctx context.Context, gameID string, userID string) error {
	// TODO: Implement game deletion logic (hard delete)
	return nil
}

// CancelGameResult contains the result of a cancel operation
type CancelGameResult struct {
	ParticipantsToNotify []models.User // List of participants to notify about cancellation
}

// CancelGame cancels a game and returns information about participants to notify
func (s *GamesService) CancelGame(ctx context.Context, gameID string, userID string) (*CancelGameResult, error) {
	logger := log.Ctx(ctx)

	// Validate UUIDs
	var gameUUID, userUUID pgtype.UUID
	if err := gameUUID.Scan(gameID); err != nil {
		return nil, &InvalidArgumentError{
			ArgumentName: "game_id",
			Message:      "invalid game ID format",
		}
	}
	if err := userUUID.Scan(userID); err != nil {
		return nil, &InvalidArgumentError{
			ArgumentName: "user_id",
			Message:      "invalid user ID format",
		}
	}

	// Get game to validate ownership and current status
	game, err := s.queries.GetGame(ctx, gameUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("game not found")
		}
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	// Verify user is the game owner
	if uuid.UUID(game.OwnerID.Bytes).String() != userID {
		return nil, ErrNotOwner
	}

	// Check if already cancelled (idempotent)
	if game.Status == string(models.GameStatusCancelled) {
		logger.Info().Msg("Game already cancelled (idempotent)")
		return &CancelGameResult{ParticipantsToNotify: []models.User{}}, nil
	}

	// Check if game has already completed
	if game.Status == string(models.GameStatusCompleted) {
		return nil, fmt.Errorf("cannot cancel completed game")
	}

	// Check if game has already started or finished
	now := time.Now()
	if now.After(game.StartTime.Time) {
		// Game has started
		gameEndTime := game.StartTime.Time.Add(time.Duration(game.DurationMinutes) * time.Minute)
		if now.After(gameEndTime) {
			return nil, ErrGameFinished
		}
		return nil, ErrGameAlreadyStarted
	}

	// Get all participants to notify (before cancelling)
	participants, err := s.queries.ListParticipantsByGame(ctx, gameUUID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get participants for notification")
		// Don't fail the cancel operation, just log the error
	}

	// Cancel the game
	_, err = s.queries.CancelGame(ctx, gameUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel game: %w", err)
	}

	logger.Info().Msg("Game cancelled successfully")

	// Prepare notification list
	result := &CancelGameResult{
		ParticipantsToNotify: make([]models.User, 0, len(participants)),
	}

	for _, p := range participants {
		result.ParticipantsToNotify = append(result.ParticipantsToNotify, models.User{
			ID:        uuid.UUID(p.UserID.Bytes).String(),
			Email:     p.Email,
			FirstName: p.FirstName,
			LastName:  p.LastName,
		})
	}

	logger.Info().Int("participantCount", len(result.ParticipantsToNotify)).Msg("Participants to notify about cancellation")

	return result, nil
}

// addOrUpdateParticipant adds a new participant or reactivates an inactive one within a transaction
func (s *GamesService) addOrUpdateParticipant(ctx context.Context, gameUUID, userUUID pgtype.UUID) error {
	// Start a transaction with row-level locking
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Use transaction-scoped queries
	txQueries := repository.New(tx).WithTx(tx)

	// Get game with row-level lock to prevent race conditions
	// This will BLOCK if another transaction has the lock, waiting until it's released
	game, err := txQueries.GetGameForUpdate(ctx, gameUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("timed out waiting for game lock - please try again")
		}
		return fmt.Errorf("failed to get game: %w", err)
	}

	// Validate game hasn't finished
	gameEndTime := game.StartTime.Time.Add(time.Duration(game.DurationMinutes) * time.Minute)
	if time.Now().After(gameEndTime) {
		return fmt.Errorf("cannot join game: game has already finished")
	}

	// Get all participants to determine status
	existingParticipants, err := txQueries.ListParticipantsByGame(ctx, gameUUID)
	if err != nil {
		return fmt.Errorf("failed to retrieve existing participants: %w", err)
	}

	// Find existing participant record and count active participants
	var existingParticipantRecord *repository.ListParticipantsByGameRow
	activeParticipants := 0
	for _, participant := range existingParticipants {
		if !InactiveParticipantStates[participant.Status] {
			activeParticipants++
		}
		if participant.UserID == userUUID {
			existingParticipantRecord = &participant
			// Don't break - we need to count all active participants
		}
	}

	// Determine participant status based on count of ACTIVE participants
	participantStatus := models.ParticipantStatusConfirmed
	if activeParticipants >= int(game.MaxParticipants) {
		participantStatus = models.ParticipantStatusWaitlist
	}

	if existingParticipantRecord == nil {
		// Create new participant
		_, err = txQueries.CreateParticipant(ctx, repository.CreateParticipantParams{
			GameID: gameUUID,
			UserID: userUUID,
			Status: string(participantStatus),
		})
		if err != nil {
			return fmt.Errorf("failed to create participant: %w", err)
		}
	} else {
		// Check if participant is in an inactive state
		if InactiveParticipantStates[existingParticipantRecord.Status] {
			// Re-joining: reset joined_at to put them at the back of the line
			_, err = txQueries.UpdateParticipantStatusResetJoinedAt(ctx, repository.UpdateParticipantStatusResetJoinedAtParams{
				ID:     existingParticipantRecord.ID,
				Status: string(participantStatus),
			})
			if err != nil {
				return fmt.Errorf("failed to update participant for rejoin: %w", err)
			}
		}
		// else: already active, nothing to do (idempotent)
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// reconcileParticipantStatuses updates participant statuses in batch to match their actual positions
// Only updates records where the status doesn't match (confirmed->waitlist or waitlist->confirmed)
func (s *GamesService) reconcileParticipantStatuses(ctx context.Context, gameUUID pgtype.UUID, maxParticipants int32) error {
	// First, check if any updates are needed (without locking)
	participants, err := s.queries.ListParticipantsByGame(ctx, gameUUID)
	if err != nil {
		return fmt.Errorf("failed to list participants: %w", err)
	}

	// Build lists of IDs that need updating
	var toConfirm []pgtype.UUID   // Waitlisted participants who should be confirmed
	var toWaitlist []pgtype.UUID  // Confirmed participants who should be waitlisted

	activeCount := 0
	for _, p := range participants {
		// Skip inactive participants
		if InactiveParticipantStates[p.Status] {
			continue
		}

		// Determine what status should be based on position
		shouldBeConfirmed := activeCount < int(maxParticipants)
		isConfirmed := p.Status == string(models.ParticipantStatusConfirmed)

		// Check if status needs updating
		if shouldBeConfirmed && !isConfirmed {
			toConfirm = append(toConfirm, p.ID)
		} else if !shouldBeConfirmed && isConfirmed {
			toWaitlist = append(toWaitlist, p.ID)
		}

		activeCount++
	}

	// If no updates needed, return early (no lock acquired)
	if len(toConfirm) == 0 && len(toWaitlist) == 0 {
		return nil
	}

	// Updates needed - start transaction with lock
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	txQueries := repository.New(tx).WithTx(tx)

	// Lock the game to prevent concurrent modifications during reconciliation
	_, err = txQueries.GetGameForUpdate(ctx, gameUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("timed out waiting for game lock during reconciliation")
		}
		return fmt.Errorf("failed to lock game for reconciliation: %w", err)
	}

	// Batch update to confirmed (only if there are records to update)
	if len(toConfirm) > 0 {
		err = txQueries.BatchUpdateParticipantsToConfirmed(ctx, toConfirm)
		if err != nil {
			return fmt.Errorf("failed to batch update participants to confirmed: %w", err)
		}
	}

	// Batch update to waitlist (only if there are records to update)
	if len(toWaitlist) > 0 {
		err = txQueries.BatchUpdateParticipantsToWaitlist(ctx, toWaitlist)
		if err != nil {
			return fmt.Errorf("failed to batch update participants to waitlist: %w", err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// JoinGame adds a user as a participant to a game and returns all participants with computed status
func (s *GamesService) JoinGame(ctx context.Context, gameID string, userID string) ([]models.Participant, error) {
	// Validate game and user UUID
	var gameUUID, userUUID pgtype.UUID
	if err := gameUUID.Scan(gameID); err != nil {
		return nil, &InvalidArgumentError{
			ArgumentName: "game_id",
			Message:      "invalid game ID format",
		}
	}
	if err := userUUID.Scan(userID); err != nil {
		return nil, &InvalidArgumentError{
			ArgumentName: "user_id",
			Message:      "invalid user ID format",
		}
	}

	// Step 1: Add or update the participant (in transaction with row lock)
	if err := s.addOrUpdateParticipant(ctx, gameUUID, userUUID); err != nil {
		return nil, err
	}

	// Step 2: Get game info for max participants
	game, err := s.queries.GetGame(ctx, gameUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	// Step 3: Reconcile all participant statuses to ensure they're accurate
	if err := s.reconcileParticipantStatuses(ctx, gameUUID, game.MaxParticipants); err != nil {
		return nil, err
	}

	// Step 4: Get final participant list - statuses are now accurate from reconciliation
	participants, err := s.queries.ListParticipantsByGame(ctx, gameUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to list participants: %w", err)
	}

	// Convert to models, using the status directly from the database (already reconciled)
	result := make([]models.Participant, 0, len(participants))
	waitlistCount := 0
	for _, p := range participants {
		// Skip inactive participants
		if InactiveParticipantStates[p.Status] {
			continue
		}

		// Use the status from the database (already correct after reconciliation)
		status := models.ParticipantStatus(p.Status)
		var waitlistPosition *int

		// Calculate waitlist position for waitlisted participants
		if status == models.ParticipantStatusWaitlist {
			waitlistCount++
			waitlistPosition = &waitlistCount
		}

		result = append(result, *convertParticipantRowToModelWithStatus(p, status, waitlistPosition))
	}

	return result, nil
}

// DropGameResult contains the result of a drop operation
type DropGameResult struct {
	PromotedUser *models.User // User promoted from waitlist (nil if no promotion)
}

// DropGame marks a user as dropped from a game and returns information about any waitlist promotions
func (s *GamesService) DropGame(ctx context.Context, gameID string, userID string) (*DropGameResult, error) {
	logger := log.Ctx(ctx)

	// Validate game and user UUID
	var gameUUID, userUUID pgtype.UUID
	if err := gameUUID.Scan(gameID); err != nil {
		return nil, &InvalidArgumentError{
			ArgumentName: "game_id",
			Message:      "invalid game ID format",
		}
	}
	if err := userUUID.Scan(userID); err != nil {
		return nil, &InvalidArgumentError{
			ArgumentName: "user_id",
			Message:      "invalid user ID format",
		}
	}

	// Get game to validate
	game, err := s.queries.GetGame(ctx, gameUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	now := time.Now()

	// Validate game hasn't finished (start time + duration > now)
	gameEndTime := game.StartTime.Time.Add(time.Duration(game.DurationMinutes) * time.Minute)
	if now.After(gameEndTime) {
		return nil, ErrGameFinished
	}

	// Validate drop deadline if one is set
	if game.DropDeadline.Valid && now.After(game.DropDeadline.Time) {
		return nil, ErrTooLate
	}

	// Get the participant record
	participant, err := s.queries.GetParticipantByGameAndUser(ctx, repository.GetParticipantByGameAndUserParams{
		GameID: gameUUID,
		UserID: userUUID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotParticipant
		}
		return nil, fmt.Errorf("failed to get participant: %w", err)
	}

	// Check if already dropped (idempotent)
	if participant.Status == string(models.ParticipantStatusDropped) {
		logger.Info().Msg("User already dropped from game (idempotent)")
		return &DropGameResult{PromotedUser: nil}, nil
	}

	// Check if there's a waitlist before dropping (to detect promotions)
	participantsBefore, err := s.queries.ListParticipantsByGame(ctx, gameUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to list participants before drop: %w", err)
	}
	hasWaitlist := len(participantsBefore) > int(game.MaxParticipants)

	// Update participant status to dropped
	_, err = s.queries.UpdateParticipantStatus(ctx, repository.UpdateParticipantStatusParams{
		ID:     participant.ID,
		Status: string(models.ParticipantStatusDropped),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update participant status: %w", err)
	}

	logger.Info().Msg("User dropped from game successfully")

	result := &DropGameResult{PromotedUser: nil}

	// Check for promotion if there was a waitlist
	if hasWaitlist {
		// Get new participant list after drop
		participantsAfter, err := s.queries.ListParticipantsByGame(ctx, gameUUID)
		if err != nil {
			// Don't fail the drop operation, just log the error
			logger.Error().Err(err).Msg("Failed to get participants after drop for promotion detection")
			return result, nil
		}

		// If we still have max_participants or more, the person at position max_participants was promoted
		if len(participantsAfter) >= int(game.MaxParticipants) {
			// The person at index (max_participants - 1) is the last confirmed person (was first on waitlist)
			promotedParticipant := participantsAfter[game.MaxParticipants-1]
			result.PromotedUser = &models.User{
				ID:        uuid.UUID(promotedParticipant.UserID.Bytes).String(),
				Email:     promotedParticipant.Email,
				FirstName: promotedParticipant.FirstName,
				LastName:  promotedParticipant.LastName,
			}
			logger.Info().
				Str("promotedUserId", result.PromotedUser.ID).
				Str("promotedUserEmail", result.PromotedUser.Email).
				Msg("User promoted from waitlist")
		}
	}

	return result, nil
}

// convertParticipantRowToModelWithStatus converts a repository.ListParticipantsByGameRow to a models.Participant with computed status
func convertParticipantRowToModelWithStatus(p repository.ListParticipantsByGameRow, status models.ParticipantStatus, waitlistPosition *int) *models.Participant {
	var teamID *string
	if p.TeamID.Valid {
		id := uuid.UUID(p.TeamID.Bytes).String()
		teamID = &id
	}

	var paymentCents *int
	if p.PaymentAmountCents.Valid {
		cents := int(p.PaymentAmountCents.Int32)
		paymentCents = &cents
	}

	return &models.Participant{
		User: models.User{
			ID:        uuid.UUID(p.UserID.Bytes).String(),
			Email:     p.Email,
			FirstName: p.FirstName,
			LastName:  p.LastName,
		},
		TeamID:             teamID,
		Status:             status,
		WaitlistPosition:   waitlistPosition,
		Paid:               p.Paid,
		PaymentAmountCents: paymentCents,
		Notes:              pgTextToStringPtr(p.Notes),
		JoinedAt:           p.JoinedAt.Time.UTC(),
		UpdatedAt:          p.UpdatedAt.Time.UTC(),
	}
}
