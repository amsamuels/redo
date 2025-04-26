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

func (s *LinkService) ListLinks(ctx context.Context, userID string) ([]model.Link, error) {
	query := `
        SELECT slug, destination, created_at
        FROM links
        WHERE user_id = $1
        ORDER BY created_at DESC
    `
	rows, err := s.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []model.Link
	for rows.Next() {
		var link model.Link
		if err := rows.Scan(&link.Slug, &link.Destination, &link.CreatedAt); err != nil {
			return nil, err
		}
		links = append(links, link)
	}

	return links, nil
}

func (s *LinkService) ResolveLink(ctx context.Context, slug string) (string, error) {
	var destination string

	query := `SELECT destination FROM links WHERE slug = $1`
	err := s.DB.QueryRowContext(ctx, query, slug).Scan(&destination)
	if err != nil {
		return "", err
	}

	return destination, nil
}

func (s *LinkService) TrackClick(ctx context.Context, slug, ip, referrer, userAgent string) error {
	query := `
		INSERT INTO clicks (id, link_id, ip, referrer, user_agent, created_at)
		SELECT gen_random_uuid(), l.id, $2, $3, $4, now()
		FROM links l
		WHERE l.slug = $1
	`
	_, err := s.DB.ExecContext(ctx, query, slug, ip, referrer, userAgent)
	return err
}

func (s *LinkService) GetClickCount(ctx context.Context, slug string) (int, error) {
	var count int

	query := `
		SELECT COUNT(*)
		FROM clicks c
		JOIN links l ON c.link_id = l.id
		WHERE l.slug = $1
	`
	err := s.DB.QueryRowContext(ctx, query, slug).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
