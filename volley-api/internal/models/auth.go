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
// For mobile clients, the token is included in the JSON response
// For web clients, the token is set in an HTTP-only secure cookie
type AuthResponse struct {
	Token *string `json:"token,omitempty"` // JWT token (only for mobile clients)
	User  User    `json:"user"`            // User details
}
