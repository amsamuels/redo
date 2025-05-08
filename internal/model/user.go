package model

type SignUpRequest struct {
	Email string `json:"email"`
}

type User struct {
	UserID    string `json:"id"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}
