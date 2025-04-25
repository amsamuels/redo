package service

import (
	"context"
	"database/sql"
	"time"

	"redo.ai/internal/model"
)

type UserService struct {
	DB *sql.DB
}

func (s *UserService) SignUp(ctx context.Context, req model.SignUpRequest) error {
	query := `
        INSERT INTO users (id, email, name, business_name, created_at)
        VALUES (gen_random_uuid(), $1, $2, $3, $4)
    `
	_, err := s.DB.ExecContext(ctx, query, req.Email, req.Name, req.BusinessName, time.Now())
	return err
}

func (s *UserService) Login(ctx context.Context, req model.LoginRequest) (*model.User, error) {
	var user model.User

	query := `
        SELECT id, email, name, business_name, created_at
        FROM users
        WHERE email = $1
    `
	err := s.DB.QueryRowContext(ctx, query, req.Email).
		Scan(&user.ID, &user.Email, &user.Name, &user.BusinessName, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
