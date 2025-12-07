package ifaces

import (
	"context"

	"github.com/gabe-dev-svc/volley/internal/repository"
	"github.com/jackc/pgx/v5/pgtype"
)

type IFaceTest interface {
	Test() error
}

// Querier is the interface for database queries
type Querier interface {
	BatchUpdateParticipantsToConfirmed(ctx context.Context, participantIds []pgtype.UUID) error
	BatchUpdateParticipantsToWaitlist(ctx context.Context, participantIds []pgtype.UUID) error
	CancelGame(ctx context.Context, id pgtype.UUID) (pgtype.UUID, error)
	CountConfirmedParticipants(ctx context.Context, gameID pgtype.UUID) (int64, error)
	CountWaitlistParticipants(ctx context.Context, gameID pgtype.UUID) (int64, error)
	CreateGame(ctx context.Context, arg repository.CreateGameParams) (repository.CreateGameRow, error)
	CreateParticipant(ctx context.Context, arg repository.CreateParticipantParams) (repository.Participant, error)
	CreateTeam(ctx context.Context, arg repository.CreateTeamParams) (repository.Team, error)
	CreateUser(ctx context.Context, arg repository.CreateUserParams) (repository.User, error)
	DeleteGame(ctx context.Context, id pgtype.UUID) error
	DeleteParticipant(ctx context.Context, id pgtype.UUID) error
	DeleteTeam(ctx context.Context, id pgtype.UUID) error
	DeleteUser(ctx context.Context, id pgtype.UUID) error
	GetGame(ctx context.Context, id pgtype.UUID) (repository.GetGameRow, error)
	GetGameForUpdate(ctx context.Context, id pgtype.UUID) (repository.GetGameForUpdateRow, error)
	GetParticipant(ctx context.Context, id pgtype.UUID) (repository.Participant, error)
	GetParticipantByGameAndUser(ctx context.Context, arg repository.GetParticipantByGameAndUserParams) (repository.Participant, error)
	GetTeam(ctx context.Context, id pgtype.UUID) (repository.Team, error)
	GetUserByEmail(ctx context.Context, email string) (repository.User, error)
	GetUserByID(ctx context.Context, id pgtype.UUID) (repository.User, error)
	ListGamesInRadius(ctx context.Context, arg repository.ListGamesInRadiusParams) ([]repository.ListGamesInRadiusRow, error)
	ListActiveParticipantsByGame(ctx context.Context, gameID pgtype.UUID) ([]repository.ListActiveParticipantsByGameRow, error)
	ListParticipantsByGame(ctx context.Context, gameID pgtype.UUID) ([]repository.ListParticipantsByGameRow, error)
	ListParticipantsByUser(ctx context.Context, userID pgtype.UUID) ([]repository.Participant, error)
	ListTeamsByGame(ctx context.Context, gameID pgtype.UUID) ([]repository.Team, error)
	UpdateGame(ctx context.Context, arg repository.UpdateGameParams) (pgtype.UUID, error)
	UpdateParticipantPayment(ctx context.Context, arg repository.UpdateParticipantPaymentParams) (repository.Participant, error)
	UpdateParticipantStatus(ctx context.Context, arg repository.UpdateParticipantStatusParams) (repository.Participant, error)
	UpdateParticipantStatusResetJoinedAt(ctx context.Context, arg repository.UpdateParticipantStatusResetJoinedAtParams) (repository.Participant, error)
	UpdateParticipantTeam(ctx context.Context, arg repository.UpdateParticipantTeamParams) (repository.Participant, error)
	UpdateUser(ctx context.Context, arg repository.UpdateUserParams) (repository.User, error)
}
