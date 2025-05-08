package user

import (
	"context"
	"database/sql"
	"time"

	"redo.ai/internal/model"
)

// UserService defines the interface for user-related operations.
type UserService interface {
	SignUp(context.Context, string, string) (*model.User, error)
	GetByID(ctx context.Context, auth0Sub string) (*model.User, error)
}

// Concrete implementation of UserService.
type UserSvc struct {
	DB *sql.DB
}

// SignUp creates a new user with only the Auth0 sub (no PII).
func (s *UserSvc) SignUp(ctx context.Context, auth0Sub, email string) (*model.User, error) {
	var user model.User
	var createdAt time.Time
	var tmpAuth0Sub string // Placeholder for auth0_sub

	query := `
        INSERT INTO users (auth0_sub, email)
        VALUES ($1, $2)
    `
	err := s.DB.QueryRowContext(ctx, query, auth0Sub).Scan(
		&user.UserID, // string
		&tmpAuth0Sub, // ignored
		&user.Role,   // string
		&createdAt,   // time.Time
	)

	if err != nil {
		return nil, err
	}

	user.CreatedAt = createdAt.Format(time.RFC3339)
	return &user, nil
}

// GetByID retrieves a user by Auth0 sub and returns the user with string ID.
func (s *UserSvc) GetByID(ctx context.Context, auth0Sub string) (*model.User, error) {
	var user model.User
	var createdAt time.Time
	var tmpAuth0Sub string // Placeholder for auth0_sub

	query := `
        SELECT id::text, auth0_sub, role, created_at
        FROM users
        WHERE auth0_sub = $1
    `

	err := s.DB.QueryRowContext(ctx, query, auth0Sub).Scan(
		&user.UserID,
		&tmpAuth0Sub, // Ignored
		&user.Role,
		&createdAt,
	)

	if err != nil {
		return nil, err
	}

	user.CreatedAt = createdAt.Format(time.RFC3339)
	return &user, nil
}
