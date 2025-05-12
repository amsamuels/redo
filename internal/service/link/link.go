// internal/service/link.go
package link

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"redo.ai/internal/model"
	"redo.ai/internal/service/user"
	"redo.ai/logger"
)

// LinkService defines the interface for link-related operations.
type LinkService interface {
	CreateLink(ctx context.Context, userID string, req model.CreateLinkRequest) error
	ListLinks(ctx context.Context, userID string) ([]model.Link, error)
	ResolveLink(ctx context.Context, slug string) (string, string, error)
	TrackClick(ctx context.Context, slug, ip, referrer, userAgent string) error
	GetClickCount(ctx context.Context, slug string) (int, error)
	DeleteLink(ctx context.Context, linkID string) error
}

type LinkSvc struct {
	DB          *sql.DB
	UserService user.UserService
}

func (s *LinkSvc) CreateLink(ctx context.Context, userID string, req model.CreateLinkRequest) error {
	query := `
        INSERT INTO links (id, user_id, slug, destination, created_at)
        VALUES (gen_random_uuid(), $1, $2, $3, $4)
    `
	_, err := s.DB.ExecContext(ctx, query, userID, req.Slug, req.Destination, time.Now().UTC().Format(time.RFC3339Nano))
	return err
}

func (s *LinkSvc) ListLinks(ctx context.Context, userID string) ([]model.Link, error) {
	var links []model.Link = make([]model.Link, 0)
	query := `
       	SELECT id::text, slug, destination, created_at
        FROM links
        WHERE user_id = $1
        ORDER BY created_at DESC
    `
	rows, err := s.DB.QueryContext(ctx, query, userID)
	if err != nil {
		logger.Error("query failed in link service listlink")
		return links, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var link model.Link
		if err := rows.Scan(&link.LinkID, &link.Slug, &link.Destination, &link.CreatedAt); err != nil {
			logger.Error("scan failded err:%s", err)
			return links, fmt.Errorf("scan failed: %w", err)
		}
		links = append(links, link)
	}

	if err := rows.Err(); err != nil {
		logger.Error("rows error: [%s]in link service listlink", err)
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return links, nil
}

func (s *LinkSvc) ResolveLink(ctx context.Context, slug string) (string, string, error) {
	var (
		linkID      string
		destination string
	)

	query := `SELECT link_id, destination FROM links WHERE slug = $1`
	err := s.DB.QueryRowContext(ctx, query, slug).Scan(&linkID, &destination)
	if err != nil {
		return "", "", err
	}

	return destination, linkID, nil
}

func (s *LinkSvc) TrackClick(ctx context.Context, slug, ip, referrer, userAgent string) error {
	query := `
		INSERT INTO clicks (id, link_id, ip, referrer, user_agent, created_at)
		SELECT gen_random_uuid(), l.id, $2, $3, $4, now()
		FROM links l
		WHERE l.slug = $1
	`
	_, err := s.DB.ExecContext(ctx, query, slug, ip, referrer, userAgent)
	return err
}

func (s *LinkSvc) GetClickCount(ctx context.Context, slug string) (int, error) {
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

func (s *LinkSvc) DeleteLink(ctx context.Context, linkID string) error {
	// Optional: check if link exists (can be skipped if DELETE handles it silently)
	var exists bool
	var err error
	checkQuery := `SELECT EXISTS(SELECT 1 FROM links WHERE id = $1)`
	if err := s.DB.QueryRowContext(ctx, checkQuery, linkID).Scan(&exists); err != nil || !exists {
		return fmt.Errorf("error: [%s] with link with ID %s does not exist", err, linkID)
	}

	// Proceed to delete
	deleteQuery := `DELETE FROM links WHERE id = $1`

	_, err = s.DB.ExecContext(ctx, deleteQuery, linkID)
	if err != nil {
		return fmt.Errorf("failed to delete link: %w", err)
	}

	return nil
}
