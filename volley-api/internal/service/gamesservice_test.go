package service

// Unit tests for GamesService using testify/mock for mocking database operations.
// These tests focus on the DropGame waitlist promotion logic.
//
// Test Coverage:
// - TestDropGame_WaitlistPromotion: Table-driven tests for promotion scenarios
// - TestDropGame_Errors: Error condition handling
//
// Coverage: 81% of DropGame function

import (
	"context"
	"testing"
	"time"

	"github.com/gabe-dev-svc/volley/internal/models"
	"github.com/gabe-dev-svc/volley/internal/repository"
	"github.com/gabe-dev-svc/volley/mocks"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Helper functions to create test data
func createTestUUID(t *testing.T, id string) pgtype.UUID {
	t.Helper()
	u, err := uuid.Parse(id)
	require.NoError(t, err)
	return pgtype.UUID{
		Bytes: u,
		Valid: true,
	}
}

func createTestParticipant(userID, email, firstName, lastName string, joinedAt time.Time) repository.ListParticipantsByGameRow {
	u, _ := uuid.Parse(userID)

	return repository.ListParticipantsByGameRow{
		UserID: pgtype.UUID{
			Bytes: u,
			Valid: true,
		},
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		JoinedAt:  pgtype.Timestamptz{Time: joinedAt, Valid: true},
	}
}

// TestDropGame_WaitlistPromotion tests the waitlist promotion logic using table-driven tests
func TestDropGame_WaitlistPromotion(t *testing.T) {
	now := time.Now()
	futureTime := now.Add(24 * time.Hour)

	gameID := "00000000-0000-0000-0000-000000000001"
	userID := "00000000-0000-0000-0000-000000000002"

	gameUUID := createTestUUID(t, gameID)
	userUUID := createTestUUID(t, userID)
	participantID := createTestUUID(t, "00000000-0000-0000-0000-000000000010")

	tests := []struct {
		name                   string
		maxParticipants        int32
		participantsBefore     []repository.ListParticipantsByGameRow
		participantsAfter      []repository.ListParticipantsByGameRow
		droppingUserStatus     string
		expectPromotion        bool
		expectedPromotedUserID string
		expectedError          error
	}{
		{
			name:            "Game at capacity with waitlist - promotion occurs",
			maxParticipants: 2,
			participantsBefore: []repository.ListParticipantsByGameRow{
				createTestParticipant("00000000-0000-0000-0000-000000000002", "user1@test.com", "User", "One", now.Add(-3*time.Hour)),
				createTestParticipant("00000000-0000-0000-0000-000000000004", "user2@test.com", "User", "Two", now.Add(-2*time.Hour)),
				createTestParticipant("00000000-0000-0000-0000-000000000003", "waitlist@test.com", "Waitlist", "User", now.Add(-1*time.Hour)),
			},
			participantsAfter: []repository.ListParticipantsByGameRow{
				createTestParticipant("00000000-0000-0000-0000-000000000004", "user2@test.com", "User", "Two", now.Add(-2*time.Hour)),
				createTestParticipant("00000000-0000-0000-0000-000000000003", "waitlist@test.com", "Waitlist", "User", now.Add(-1*time.Hour)),
			},
			droppingUserStatus:     string(models.ParticipantStatusConfirmed),
			expectPromotion:        true,
			expectedPromotedUserID: "00000000-0000-0000-0000-000000000003",
			expectedError:          nil,
		},
		{
			name:            "Game at capacity, no waitlist - no promotion",
			maxParticipants: 2,
			participantsBefore: []repository.ListParticipantsByGameRow{
				createTestParticipant("00000000-0000-0000-0000-000000000002", "user1@test.com", "User", "One", now.Add(-3*time.Hour)),
				createTestParticipant("00000000-0000-0000-0000-000000000004", "user2@test.com", "User", "Two", now.Add(-2*time.Hour)),
			},
			participantsAfter: []repository.ListParticipantsByGameRow{
				createTestParticipant("00000000-0000-0000-0000-000000000004", "user2@test.com", "User", "Two", now.Add(-2*time.Hour)),
			},
			droppingUserStatus: string(models.ParticipantStatusConfirmed),
			expectPromotion:    false,
			expectedError:      nil,
		},
		{
			name:            "Game below capacity - no promotion",
			maxParticipants: 5,
			participantsBefore: []repository.ListParticipantsByGameRow{
				createTestParticipant("00000000-0000-0000-0000-000000000002", "user1@test.com", "User", "One", now.Add(-3*time.Hour)),
				createTestParticipant("00000000-0000-0000-0000-000000000004", "user2@test.com", "User", "Two", now.Add(-2*time.Hour)),
			},
			participantsAfter: []repository.ListParticipantsByGameRow{
				createTestParticipant("00000000-0000-0000-0000-000000000004", "user2@test.com", "User", "Two", now.Add(-2*time.Hour)),
			},
			droppingUserStatus: string(models.ParticipantStatusConfirmed),
			expectPromotion:    false,
			expectedError:      nil,
		},
		{
			name:            "Multiple waitlisted users - first one gets promoted",
			maxParticipants: 2,
			participantsBefore: []repository.ListParticipantsByGameRow{
				createTestParticipant("00000000-0000-0000-0000-000000000002", "user1@test.com", "User", "One", now.Add(-5*time.Hour)),
				createTestParticipant("00000000-0000-0000-0000-000000000004", "user2@test.com", "User", "Two", now.Add(-4*time.Hour)),
				createTestParticipant("00000000-0000-0000-0000-000000000003", "waitlist1@test.com", "Waitlist", "One", now.Add(-3*time.Hour)),
				createTestParticipant("00000000-0000-0000-0000-000000000005", "waitlist2@test.com", "Waitlist", "Two", now.Add(-2*time.Hour)),
			},
			participantsAfter: []repository.ListParticipantsByGameRow{
				createTestParticipant("00000000-0000-0000-0000-000000000004", "user2@test.com", "User", "Two", now.Add(-4*time.Hour)),
				createTestParticipant("00000000-0000-0000-0000-000000000003", "waitlist1@test.com", "Waitlist", "One", now.Add(-3*time.Hour)),
				createTestParticipant("00000000-0000-0000-0000-000000000005", "waitlist2@test.com", "Waitlist", "Two", now.Add(-2*time.Hour)),
			},
			droppingUserStatus:     string(models.ParticipantStatusConfirmed),
			expectPromotion:        true,
			expectedPromotedUserID: "00000000-0000-0000-0000-000000000003",
			expectedError:          nil,
		},
		{
			name:            "User already dropped - idempotent, no promotion",
			maxParticipants: 2,
			participantsBefore: []repository.ListParticipantsByGameRow{
				createTestParticipant("00000000-0000-0000-0000-000000000004", "user2@test.com", "User", "Two", now.Add(-2*time.Hour)),
				createTestParticipant("00000000-0000-0000-0000-000000000003", "waitlist@test.com", "Waitlist", "User", now.Add(-1*time.Hour)),
			},
			participantsAfter:  nil, // Won't be called since user already dropped
			droppingUserStatus: string(models.ParticipantStatusDropped),
			expectPromotion:    false,
			expectedError:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockQuerier := mocks.NewQuerier(t)
			service := &GamesService{queries: mockQuerier, pool: nil}
			ctx := context.Background()

			// Mock GetGame
			mockQuerier.On("GetGame", ctx, gameUUID).Return(repository.GetGameRow{
				ID:              gameUUID,
				MaxParticipants: tt.maxParticipants,
				StartTime:       pgtype.Timestamptz{Time: futureTime, Valid: true},
				DurationMinutes: 90,
				DropDeadline:    pgtype.Timestamptz{Valid: false},
			}, nil)

			// Mock GetParticipantByGameAndUser
			mockQuerier.On("GetParticipantByGameAndUser", ctx, repository.GetParticipantByGameAndUserParams{
				GameID: gameUUID,
				UserID: userUUID,
			}).Return(repository.Participant{
				ID:     participantID,
				Status: tt.droppingUserStatus,
			}, nil)

			// If user already dropped, return early
			if tt.droppingUserStatus == string(models.ParticipantStatusDropped) {
				// No more mocks needed
			} else {
				// Mock ListParticipantsByGame (before drop)
				mockQuerier.On("ListParticipantsByGame", ctx, gameUUID).Return(tt.participantsBefore, nil).Once()

				// Mock UpdateParticipantStatus
				mockQuerier.On("UpdateParticipantStatus", ctx, repository.UpdateParticipantStatusParams{
					ID:     participantID,
					Status: string(models.ParticipantStatusDropped),
				}).Return(repository.Participant{}, nil)

				// Mock ListParticipantsByGame (after drop) - only if there was a waitlist
				if len(tt.participantsBefore) > int(tt.maxParticipants) {
					mockQuerier.On("ListParticipantsByGame", ctx, gameUUID).Return(tt.participantsAfter, nil).Once()
				}
			}

			// Execute
			result, err := service.DropGame(ctx, gameID, userID)

			// Assert
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)

				if tt.expectPromotion {
					require.NotNil(t, result.PromotedUser, "Expected promotion but got nil")
					assert.Equal(t, tt.expectedPromotedUserID, result.PromotedUser.ID)
				} else {
					assert.Nil(t, result.PromotedUser, "Expected no promotion but got one")
				}
			}

			// Verify all expectations were met
			mockQuerier.AssertExpectations(t)
		})
	}
}

// TestDropGame_Errors tests error conditions
func TestDropGame_Errors(t *testing.T) {
	now := time.Now()
	pastTime := now.Add(-2 * time.Hour)
	futureTime := now.Add(24 * time.Hour)

	gameID := "00000000-0000-0000-0000-000000000001"
	userID := "00000000-0000-0000-0000-000000000002"

	gameUUID := createTestUUID(t, gameID)

	tests := []struct {
		name          string
		setupMocks    func(*mocks.Querier)
		expectedError error
	}{
		{
			name: "Game has finished",
			setupMocks: func(m *mocks.Querier) {
				m.On("GetGame", mock.Anything, gameUUID).Return(repository.GetGameRow{
					ID:              gameUUID,
					MaxParticipants: 10,
					StartTime:       pgtype.Timestamptz{Time: pastTime, Valid: true},
					DurationMinutes: 60,
					DropDeadline:    pgtype.Timestamptz{Valid: false},
				}, nil)
			},
			expectedError: ErrGameFinished,
		},
		{
			name: "Drop deadline passed",
			setupMocks: func(m *mocks.Querier) {
				m.On("GetGame", mock.Anything, gameUUID).Return(repository.GetGameRow{
					ID:              gameUUID,
					MaxParticipants: 10,
					StartTime:       pgtype.Timestamptz{Time: futureTime, Valid: true},
					DurationMinutes: 90,
					DropDeadline:    pgtype.Timestamptz{Time: pastTime, Valid: true},
				}, nil)
			},
			expectedError: ErrTooLate,
		},
		{
			name: "User not a participant",
			setupMocks: func(m *mocks.Querier) {
				m.On("GetGame", mock.Anything, gameUUID).Return(repository.GetGameRow{
					ID:              gameUUID,
					MaxParticipants: 10,
					StartTime:       pgtype.Timestamptz{Time: futureTime, Valid: true},
					DurationMinutes: 90,
					DropDeadline:    pgtype.Timestamptz{Valid: false},
				}, nil)
				m.On("GetParticipantByGameAndUser", mock.Anything, mock.Anything).Return(repository.Participant{}, pgx.ErrNoRows)
			},
			expectedError: ErrNotParticipant,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockQuerier := mocks.NewQuerier(t)
			service := &GamesService{queries: mockQuerier, pool: nil}
			ctx := context.Background()

			tt.setupMocks(mockQuerier)

			// Execute
			result, err := service.DropGame(ctx, gameID, userID)

			// Assert
			assert.ErrorIs(t, err, tt.expectedError)
			assert.Nil(t, result)

			// Verify all expectations were met
			mockQuerier.AssertExpectations(t)
		})
	}
}

// TestCancelGame_Success tests successful game cancellation scenarios
func TestCancelGame_Success(t *testing.T) {
	gameID := "550e8400-e29b-41d4-a716-446655440001"
	ownerID := "550e8400-e29b-41d4-a716-446655440002"
	gameUUID := createTestUUID(t, gameID)
	ownerUUID := createTestUUID(t, ownerID)

	tests := []struct {
		name                 string
		gameStatus           string
		startTime            time.Time
		participants         []repository.ListParticipantsByGameRow
		expectedNotifyCount  int
		expectCancelGameCall bool
		description          string
	}{
		{
			name:       "Cancel open game with participants",
			gameStatus: string(models.GameStatusOpen),
			startTime:  time.Now().Add(2 * time.Hour),
			participants: []repository.ListParticipantsByGameRow{
				createTestParticipant("550e8400-e29b-41d4-a716-446655440010", "user1@example.com", "John", "Doe", time.Now()),
				createTestParticipant("550e8400-e29b-41d4-a716-446655440011", "user2@example.com", "Jane", "Smith", time.Now()),
			},
			expectedNotifyCount:  2,
			expectCancelGameCall: true,
			description:          "Should cancel game and return all participants to notify",
		},
		{
			name:                 "Cancel game with no participants",
			gameStatus:           string(models.GameStatusOpen),
			startTime:            time.Now().Add(2 * time.Hour),
			participants:         []repository.ListParticipantsByGameRow{},
			expectedNotifyCount:  0,
			expectCancelGameCall: true,
			description:          "Should cancel game even with no participants",
		},
		{
			name:                 "Cancel already cancelled game (idempotent)",
			gameStatus:           string(models.GameStatusCancelled),
			startTime:            time.Now().Add(2 * time.Hour),
			participants:         nil,
			expectedNotifyCount:  0,
			expectCancelGameCall: false,
			description:          "Should return success without calling CancelGame",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockQuerier := mocks.NewQuerier(t)
			service := &GamesService{queries: mockQuerier, pool: nil}
			ctx := context.Background()

			// Mock GetGame
			mockQuerier.On("GetGame", ctx, gameUUID).Return(repository.GetGameRow{
				ID:              gameUUID,
				OwnerID:         ownerUUID,
				Status:          tt.gameStatus,
				StartTime:       pgtype.Timestamptz{Time: tt.startTime, Valid: true},
				DurationMinutes: 90,
			}, nil)

			// Mock ListParticipantsByGame (only if game is not already cancelled)
			if tt.expectCancelGameCall {
				mockQuerier.On("ListParticipantsByGame", ctx, gameUUID).Return(tt.participants, nil)
				mockQuerier.On("CancelGame", ctx, gameUUID).Return(gameUUID, nil)
			}

			// Execute
			result, err := service.CancelGame(ctx, gameID, ownerID)

			// Assert
			require.NoError(t, err, tt.description)
			require.NotNil(t, result)
			assert.Equal(t, tt.expectedNotifyCount, len(result.ParticipantsToNotify), "Should notify correct number of participants")

			// Verify all expectations were met
			mockQuerier.AssertExpectations(t)
		})
	}
}

// TestCancelGame_Errors tests error scenarios for game cancellation
func TestCancelGame_Errors(t *testing.T) {
	gameID := "550e8400-e29b-41d4-a716-446655440001"
	ownerID := "550e8400-e29b-41d4-a716-446655440002"
	nonOwnerID := "550e8400-e29b-41d4-a716-446655440003"
	gameUUID := createTestUUID(t, gameID)
	ownerUUID := createTestUUID(t, ownerID)

	tests := []struct {
		name          string
		gameID        string
		userID        string
		setupMocks    func(m *mocks.Querier)
		expectedError error
		description   string
	}{
		{
			name:   "Invalid game UUID",
			gameID: "invalid-uuid",
			userID: ownerID,
			setupMocks: func(m *mocks.Querier) {
				// No mocks needed - should fail validation
			},
			expectedError: nil, // Will be a wrapped error
			description:   "Should fail with invalid game UUID",
		},
		{
			name:   "Invalid user UUID",
			gameID: gameID,
			userID: "invalid-uuid",
			setupMocks: func(m *mocks.Querier) {
				// No mocks needed - should fail validation
			},
			expectedError: nil, // Will be a wrapped error
			description:   "Should fail with invalid user UUID",
		},
		{
			name:   "Game not found",
			gameID: gameID,
			userID: ownerID,
			setupMocks: func(m *mocks.Querier) {
				m.On("GetGame", mock.Anything, gameUUID).Return(repository.GetGameRow{}, pgx.ErrNoRows)
			},
			expectedError: nil, // Will be a wrapped "game not found" error
			description:   "Should fail when game does not exist",
		},
		{
			name:   "Non-owner tries to cancel",
			gameID: gameID,
			userID: nonOwnerID,
			setupMocks: func(m *mocks.Querier) {
				m.On("GetGame", mock.Anything, gameUUID).Return(repository.GetGameRow{
					ID:              gameUUID,
					OwnerID:         ownerUUID,
					Status:          string(models.GameStatusOpen),
					StartTime:       pgtype.Timestamptz{Time: time.Now().Add(2 * time.Hour), Valid: true},
					DurationMinutes: 90,
				}, nil)
			},
			expectedError: ErrNotOwner,
			description:   "Should fail when non-owner tries to cancel",
		},
		{
			name:   "Cancel completed game",
			gameID: gameID,
			userID: ownerID,
			setupMocks: func(m *mocks.Querier) {
				m.On("GetGame", mock.Anything, gameUUID).Return(repository.GetGameRow{
					ID:              gameUUID,
					OwnerID:         ownerUUID,
					Status:          string(models.GameStatusCompleted),
					StartTime:       pgtype.Timestamptz{Time: time.Now().Add(-2 * time.Hour), Valid: true},
					DurationMinutes: 90,
				}, nil)
			},
			expectedError: nil, // Will be a wrapped error about completed games
			description:   "Should fail when trying to cancel completed game",
		},
		{
			name:   "Cancel game that already started",
			gameID: gameID,
			userID: ownerID,
			setupMocks: func(m *mocks.Querier) {
				m.On("GetGame", mock.Anything, gameUUID).Return(repository.GetGameRow{
					ID:              gameUUID,
					OwnerID:         ownerUUID,
					Status:          string(models.GameStatusInProgress),
					StartTime:       pgtype.Timestamptz{Time: time.Now().Add(-30 * time.Minute), Valid: true},
					DurationMinutes: 90,
				}, nil)
			},
			expectedError: ErrGameAlreadyStarted,
			description:   "Should fail when game has already started",
		},
		{
			name:   "Cancel game that already finished",
			gameID: gameID,
			userID: ownerID,
			setupMocks: func(m *mocks.Querier) {
				m.On("GetGame", mock.Anything, gameUUID).Return(repository.GetGameRow{
					ID:              gameUUID,
					OwnerID:         ownerUUID,
					Status:          string(models.GameStatusOpen),
					StartTime:       pgtype.Timestamptz{Time: time.Now().Add(-2 * time.Hour), Valid: true},
					DurationMinutes: 90,
				}, nil)
			},
			expectedError: ErrGameFinished,
			description:   "Should fail when game has already finished",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockQuerier := mocks.NewQuerier(t)
			service := &GamesService{queries: mockQuerier, pool: nil}
			ctx := context.Background()

			tt.setupMocks(mockQuerier)

			// Execute
			result, err := service.CancelGame(ctx, tt.gameID, tt.userID)

			// Assert
			require.Error(t, err, tt.description)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			}
			assert.Nil(t, result)

			// Verify all expectations were met
			mockQuerier.AssertExpectations(t)
		})
	}
}
