// internal/service/link.go
package service

import (
	"context"
	"database/sql"
	"time"

	"redo.ai/internal/model"
)

type LinkService struct {
	DB *sql.DB
}

func (s *LinkService) CreateLink(ctx context.Context, companyID string, req model.CreateLinkRequest) error {
	query := `
        INSERT INTO links (id, company_id, slug, destination, created_at)
        VALUES (gen_random_uuid(), $1, $2, $3, $4)
    `
	_, err := s.DB.ExecContext(ctx, query, companyID, req.Slug, req.Destination, time.Now())
	return err
}

// internal/service/link.go
func (s *LinkService) ResolveLink(ctx context.Context, slug string) (string, error) {
	var destination string

	query := `SELECT destination FROM links WHERE slug = $1`
	err := s.DB.QueryRowContext(ctx, query, slug).Scan(&destination)
	if err != nil {
		return "", err
	}

	return destination, nil
}
