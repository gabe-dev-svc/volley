package models

import "time"

// GameCategory represents the sport category
type GameCategory string

const (
	GameCategorySoccer          GameCategory = "soccer"
	GameCategoryBasketball      GameCategory = "basketball"
	GameCategoryPickleball      GameCategory = "pickleball"
	GameCategoryFlagFootball    GameCategory = "flag_football"
	GameCategoryVolleyball      GameCategory = "volleyball"
	GameCategoryUltimateFrisbee GameCategory = "ultimate_frisbee"
	GameCategoryTennis          GameCategory = "tennis"
	GameCategoryOther           GameCategory = "other"
)

// GameStatus represents the current status of a game
type GameStatus string

const (
	GameStatusOpen       GameStatus = "open"        // Accepting sign-ups
	GameStatusFull       GameStatus = "full"        // Roster full, waitlist only
	GameStatusClosed     GameStatus = "closed"      // Sign-up deadline passed
	GameStatusInProgress GameStatus = "in_progress" // Game is happening
	GameStatusCompleted  GameStatus = "completed"   // Game finished
	GameStatusCancelled  GameStatus = "cancelled"   // Game cancelled
)

// SkillLevel represents the skill level requirement for a game
type SkillLevel string

const (
	SkillLevelBeginner     SkillLevel = "beginner"
	SkillLevelIntermediate SkillLevel = "intermediate"
	SkillLevelAdvanced     SkillLevel = "advanced"
	SkillLevelAll          SkillLevel = "all"
)

// PricingType represents how the game is priced
type PricingType string

const (
	PricingTypeFree      PricingType = "free"       // No cost
	PricingTypeTotal     PricingType = "total"      // Total amount split among participants
	PricingTypePerPerson PricingType = "per_person" // Fixed price per person
)

// ParticipantStatus represents the status of a participant
type ParticipantStatus string

const (
	ParticipantStatusConfirmed ParticipantStatus = "confirmed" // Confirmed participant
	ParticipantStatusWaitlist  ParticipantStatus = "waitlist"  // On waitlist
	ParticipantStatusDropped   ParticipantStatus = "dropped"   // Dropped from game
	ParticipantStatusDeclined  ParticipantStatus = "declined"  // Declined invitation
	ParticipantStatusRemoved   ParticipantStatus = "removed"   // Removed from game
)

// Location represents the location details of a game
type Location struct {
	Name      string   `json:"name"`                // Venue or field name
	Address   *string  `json:"address,omitempty"`   // Street address
	Latitude  *float64 `json:"latitude,omitempty"`  // Latitude coordinate
	Longitude *float64 `json:"longitude,omitempty"` // Longitude coordinate
	Notes     *string  `json:"notes,omitempty"`     // Additional location details
}

// Pricing represents the pricing details of a game
type Pricing struct {
	Type        PricingType `json:"type"`        // Pricing type
	AmountCents int         `json:"amountCents"` // Amount in cents (0 for free)
	Currency    string      `json:"currency"`    // Currency code (default: USD)
}

// User represents a basic user structure
type User struct {
	ID        string    `json:"id"`                  // User UUID
	Email     string    `json:"email"`               // User email
	FirstName string    `json:"firstName"`           // User first name
	LastName  string    `json:"lastName"`            // User last name
	CreatedAt time.Time `json:"createdAt,omitempty"` // Account creation timestamp
}

// Team represents a team in a game
type Team struct {
	ID        string    `json:"id"`              // Team UUID
	GameID    string    `json:"gameId"`          // Game UUID this team belongs to
	Name      string    `json:"name"`            // Team name
	Color     *string   `json:"color,omitempty"` // Hex color code
	CreatedAt time.Time `json:"createdAt"`       // Team creation timestamp
}

// Participant represents a user's participation in a game
type Participant struct {
	User                                 // Embedded user (id, email, name, createdAt)
	TeamID             *string           `json:"teamId,omitempty"`             // Team UUID (if assigned)
	Status             ParticipantStatus `json:"status"`                       // Participant status
	WaitlistPosition   *int              `json:"waitlistPosition,omitempty"`   // Position in waitlist
	Paid               bool              `json:"paid"`                         // Payment status
	PaymentAmountCents *int              `json:"paymentAmountCents,omitempty"` // Amount paid in cents
	Notes              *string           `json:"notes,omitempty"`              // Additional notes
	JoinedAt           time.Time         `json:"joinedAt"`                     // When they joined
	UpdatedAt          time.Time         `json:"updatedAt"`                    // Last update timestamp
}

// GameSummary represents essential game details for list views
type GameSummary struct {
	ID                      string             `json:"id"`                                // Game UUID
	Category                GameCategory       `json:"category"`                          // Sport category
	Title                   *string            `json:"title,omitempty"`                   // Custom title
	Description             *string            `json:"description,omitempty"`             // Game description
	Location                Location           `json:"location"`                          // Location details
	StartTime               time.Time          `json:"startTime"`                         // Game start time
	DurationMinutes         int                `json:"durationMinutes"`                   // Duration in minutes
	MaxParticipants         int                `json:"maxParticipants"`                   // Maximum number of players
	SignupCount             int                `json:"signupCount"`                       // Number of participants signed up
	Pricing                 Pricing            `json:"pricing"`                           // Pricing details
	SignupDeadline          time.Time          `json:"signupDeadline"`                    // Sign-up deadline
	SkillLevel              SkillLevel         `json:"skillLevel"`                        // Required skill level
	Status                  GameStatus         `json:"status"`                            // Current game status
	UserParticipationStatus *ParticipantStatus `json:"userParticipationStatus,omitempty"` // Current user's participation status (if authenticated)
}

// Game represents a pickup sports game with full details
type Game struct {
	ID                    string        `json:"id"`                              // Game UUID
	Owner                 *User         `json:"owner,omitempty"`                 // Owner user details
	Category              GameCategory  `json:"category"`                        // Sport category
	Title                 *string       `json:"title,omitempty"`                 // Custom title
	Description           *string       `json:"description,omitempty"`           // Game description
	Location              Location      `json:"location"`                        // Location details
	StartTime             time.Time     `json:"startTime"`                       // Game start time
	DurationMinutes       int           `json:"durationMinutes"`                 // Duration in minutes
	MaxParticipants       int           `json:"maxParticipants"`                 // Maximum number of players
	ConfirmedParticipants []Participant `json:"confirmedParticipants,omitempty"` // Confirmed participants (up to max)
	Waitlist              []Participant `json:"waitlist,omitempty"`              // Waitlisted participants (beyond max)
	Pricing               Pricing       `json:"pricing"`                         // Pricing details
	SignupDeadline        time.Time     `json:"signupDeadline"`                  // Sign-up deadline
	DropDeadline          *time.Time    `json:"dropDeadline,omitempty"`          // Drop deadline (optional)
	SkillLevel            SkillLevel    `json:"skillLevel"`                      // Required skill level
	Notes                 *string       `json:"notes,omitempty"`                 // Additional notes
	Status                GameStatus    `json:"status"`                          // Current game status
	CancelledAt           *time.Time    `json:"cancelledAt,omitempty"`           // When the game was cancelled
	CreatedAt             time.Time     `json:"createdAt"`                       // Creation timestamp
	UpdatedAt             time.Time     `json:"updatedAt"`                       // Last update timestamp
}

// CreateGameRequest represents a request to create a new game
type CreateGameRequest struct {
	Category        GameCategory `json:"category" binding:"required"`               // Sport category
	Title           *string      `json:"title,omitempty"`                           // Custom title
	Description     *string      `json:"description,omitempty"`                     // Game description
	Location        Location     `json:"location" binding:"required"`               // Location details
	StartTime       time.Time    `json:"startTime" binding:"required"`              // Game start time
	DurationMinutes int          `json:"durationMinutes" binding:"required,min=15"` // Duration in minutes
	MaxParticipants int          `json:"maxParticipants" binding:"required,min=2"`  // Maximum number of players
	Pricing         Pricing      `json:"pricing" binding:"required"`                // Pricing details
	SignupDeadline  *time.Time   `json:"signupDeadline,omitempty"`                  // Sign-up deadline (defaults to start_time)
	DropDeadline    *time.Time   `json:"dropDeadline,omitempty"`                    // Drop deadline (optional)
	SkillLevel      *SkillLevel  `json:"skillLevel,omitempty"`                      // Required skill level (defaults to "all")
	Notes           *string      `json:"notes,omitempty"`                           // Additional notes
}

// ListGamesResponse represents the response for listing games
type ListGamesResponse struct {
	Games []GameSummary `json:"games"` // List of game summaries
}

// UpdateGameRequest represents a request to update an existing game
type UpdateGameRequest struct {
	Title           *string     `json:"title,omitempty"`                                      // Custom title
	Description     *string     `json:"description,omitempty"`                                // Game description
	Location        *Location   `json:"location,omitempty"`                                   // Location details
	StartTime       *time.Time  `json:"startTime,omitempty"`                                  // Game start time
	DurationMinutes *int        `json:"durationMinutes,omitempty" binding:"omitempty,min=15"` // Duration in minutes
	MaxParticipants *int        `json:"maxParticipants,omitempty" binding:"omitempty,min=2"`  // Maximum number of players
	Pricing         *Pricing    `json:"pricing,omitempty"`                                    // Pricing details
	SignupDeadline  *time.Time  `json:"signupDeadline,omitempty"`                             // Sign-up deadline
	SkillLevel      *SkillLevel `json:"skillLevel,omitempty"`                                 // Required skill level
	Notes           *string     `json:"notes,omitempty"`                                      // Additional notes
	Status          *GameStatus `json:"status,omitempty"`                                     // Game status
}
