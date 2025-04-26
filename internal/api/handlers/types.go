package handlers

import "time"

// SignUpHandler - creates a new user with provided business name and associates to Auth0 sub (email as identifier).
type SignUpRequest struct {
	Email        string `json:"email"`
	Name         string `json:"name"`
	BusinessName string `json:"business_name"`
}

type LoginRequest struct {
	Email string `json:"email"`
}

type UserResponse struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	BusinessName string    `json:"business_name"`
	CreatedAt    time.Time `json:"created_at"`
}
