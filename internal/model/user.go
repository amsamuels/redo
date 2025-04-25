package model

type SignUpRequest struct {
	Email        string `json:"email"`
	Name         string `json:"name"`
	BusinessName string `json:"business_name"`
}

type LoginRequest struct {
	Email string `json:"email"`
}

type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	BusinessName string `json:"business_name"`
	CreatedAt    string `json:"created_at"`
}
