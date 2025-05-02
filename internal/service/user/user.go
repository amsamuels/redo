package user

import (
	"context"
	"database/sql"
	"time"

	"redo.ai/internal/model"
)

// UserService defines the interface for user-related operations.
type UserService interface {
	SignUp(context.Context, string, model.SignUpRequest) error
	GetByID(ctx context.Context, userID string) (*model.User, error)
}

// Concrete implementation of LinkService.
type UserSvc struct {
	DB *sql.DB
}

// SignUp creates a new user with ID from sub and other data from JWT.
func (s *UserSvc) SignUp(ctx context.Context, userID string, req model.SignUpRequest) error {
	query := `
        INSERT INTO users (id, email, name, business_name)
        VALUES ($1, $2, $3, $4)
    `
	_, err := s.DB.ExecContext(ctx, query, userID, req.Email, req.Name, req.BusinessName)
	return err
}

// GetByID retrieves a user by Auth0 sub (stored as ID).
func (s *UserSvc) GetByID(ctx context.Context, userID string) (*model.User, error) {
	var user model.User
	var createdAt time.Time

	query := `
        SELECT id, email, name, business_name, created_at
        FROM users
        WHERE id = $1
    `

	err := s.DB.QueryRowContext(ctx, query, userID).
		Scan(&user.ID, &user.Email, &user.Name, &user.BusinessName, &createdAt)
	if err != nil {
		return nil, err
	}

	user.CreatedAt = createdAt.Format(time.RFC3339)
	return &user, nil
}
