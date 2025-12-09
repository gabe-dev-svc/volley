package models

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	FirstName string `json:"firstName" binding:"required,min=1,max=100"`
	LastName  string `json:"lastName" binding:"required,min=1,max=100"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8,max=128"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents a successful authentication response
// For mobile clients, tokens are included in the JSON response
// For web clients, tokens are set in HTTP-only secure cookies
type AuthResponse struct {
	Token        *string `json:"token,omitempty"`        // JWT access token (included for mobile, cookie for web)
	RefreshToken *string `json:"refreshToken,omitempty"` // Refresh token (included for mobile, cookie for web)
	User         User    `json:"user"`                   // User details
}

// RefreshTokenRequest represents a request to refresh an access token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}
