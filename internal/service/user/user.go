package user

import (
	"context"
	"database/sql"

	"redo.ai/internal/model"
)

// UserService defines the interface for user-related operations.
type UserService interface {
	SignUp(context.Context, string, string) (*model.User, error)
	GetByID(ctx context.Context, auth0Sub string) (*model.User, error)
	UserExists(ctx context.Context, userID string) (bool, error)
}

// Concrete implementation of UserService.
type UserSvc struct {
	DB *sql.DB
}

// SignUp creates a new user with only the Auth0 sub (no PII).
func (s *UserSvc) SignUp(ctx context.Context, auth0Sub, email string) (*model.User, error) {
	var user model.User

	query := `
        INSERT INTO users (auth0_sub, email)
        VALUES ($1, $2)
    `
	err := s.DB.QueryRowContext(ctx, query, auth0Sub).Scan(
		&user.UserID,
		&user.Role,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserSvc) UserExists(ctx context.Context, userID string) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)"
	var exists bool
	err := s.DB.QueryRowContext(ctx, query, userID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// GetByID retrieves a user by Auth0 sub and returns the user with string ID.
func (s *UserSvc) GetByID(ctx context.Context, auth0Sub string) (*model.User, error) {
	var user model.User

	query := `
        SELECT id::text, role
        FROM users
        WHERE auth0_sub = $1
    `

	err := s.DB.QueryRowContext(ctx, query, auth0Sub).Scan(
		&user.UserID,
		&user.Role,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}
