package user

import (
	"context"
	"database/sql"
	"time"

	"redo.ai/internal/model"
)

// UserService defines the interface for user-related operations.
type UserService interface {
	SignUp(ctx context.Context, req model.SignUpRequest) error
	GetByEmail(ctx context.Context, email string) (*model.User, error)
}

// Concrete implementation of LinkService.
type UserSvc struct {
	DB *sql.DB
}

func (s *UserSvc) SignUp(ctx context.Context, req model.SignUpRequest) error {
	query := `
        INSERT INTO users (id, email, name, business_name, created_at)
        VALUES (gen_random_uuid(), $1, $2, $3, $4)
    `
	_, err := s.DB.ExecContext(ctx, query, req.Email, req.Name, req.BusinessName, time.Now())
	return err
}

// GetByEmail retrieves a user by email.
func (s *UserSvc) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User

	query := `
        SELECT id, email, name, business_name, created_at
        FROM users
        WHERE email = $1
    `
	var createdAt time.Time
	err := s.DB.QueryRowContext(ctx, query, email).
		Scan(&user.ID, &user.Email, &user.Name, &user.BusinessName, &createdAt)
	if err != nil {
		return nil, err
	}

	user.CreatedAt = createdAt.Format(time.RFC3339)
	return &user, nil
}
