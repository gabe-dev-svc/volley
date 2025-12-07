package repository

// ParticipantDetail is a type alias for participant data joined with user information.
// This allows us to use a single type across multiple queries that return the same structure.
type ParticipantDetail = ListParticipantsByGameRow

// ToParticipantDetail converts ListActiveParticipantsByGameRow to ParticipantDetail.
// Since both types have identical fields, we can safely convert between them.
func ToParticipantDetail(p ListActiveParticipantsByGameRow) ParticipantDetail {
	return ListParticipantsByGameRow{
		ID:                 p.ID,
		GameID:             p.GameID,
		UserID:             p.UserID,
		TeamID:             p.TeamID,
		Status:             p.Status,
		Paid:               p.Paid,
		PaymentAmountCents: p.PaymentAmountCents,
		Notes:              p.Notes,
		JoinedAt:           p.JoinedAt,
		UpdatedAt:          p.UpdatedAt,
		Email:              p.Email,
		FirstName:          p.FirstName,
		LastName:           p.LastName,
	}
}
