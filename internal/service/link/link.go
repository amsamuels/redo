// internal/service/link.go
package link

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"redo.ai/internal/model"
	"redo.ai/internal/service/user"
	"redo.ai/logger"
)

// LinkService defines the interface for link-related operations.
type LinkService interface {
	CreateLink(ctx context.Context, userID string, req model.CreateLinkRequest) (model.Link, error)
	ListLinks(ctx context.Context, userID string) ([]model.Link, error)
	ResolveLink(ctx context.Context, shortCode string) (string, string, error)
	ResolveUserSlug(ctx context.Context, userID string, slug string) (model.Link, error)
	TrackClick(ctx context.Context, shortCode, ip, referrer, userAgent string) error
	GetClickCount(ctx context.Context, shortCode string) (int, error)
	DeleteLink(ctx context.Context, userID, linkID string) error
}

var ErrSlugAlreadyExists = errors.New("slug already exists")
var ErrLinkNotFound = errors.New("link not found")

type LinkSvc struct {
	DB          *sql.DB
	UserService user.UserService
}

func (s *LinkSvc) CreateLink(ctx context.Context, userID string, req model.CreateLinkRequest) (model.Link, error) {
	query := `
        INSERT INTO links (user_id, slug, destination, created_at)
        VALUES ($1, $2, $3, $4)
        RETURNING id, short_code, created_at, is_active
    `
	var (
		id, shortCode string
		createdAt     time.Time
		isactive      bool
	)

	err := s.DB.QueryRowContext(
		ctx,
		query,
		userID,
		req.Slug,
		req.Destination,
		time.Now().UTC(),
	).Scan(&id, &shortCode, &createdAt, &isactive)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			if pqErr.Constraint == "unique_user_slug" {
				logger.Warn("CreateLink: duplicate slug for userID=%s: %s", userID, req.Slug)
				return model.Link{}, ErrSlugAlreadyExists
			}
		}
		logger.Error("CreateLink: failed to insert link: %v", err)
		return model.Link{}, fmt.Errorf("create link failed: %w", err)
	}

	return model.Link{
		LinkID:      id,
		Slug:        req.Slug,
		ShortCode:   shortCode,
		Destination: req.Destination,
		Is_active:   isactive,
		CreatedAt:   createdAt.Format(time.RFC3339Nano),
	}, nil
}

func (s *LinkSvc) ListLinks(ctx context.Context, userID string) ([]model.Link, error) {
	var links []model.Link = make([]model.Link, 0)
	query := `
        SELECT id::text, slug, short_code, destination, created_at, is_active
        FROM links
        WHERE user_id = $1
        ORDER BY created_at DESC
    `

	rows, err := s.DB.QueryContext(ctx, query, userID)
	if err != nil {
		logger.Error("ListLinks: query failed: %v", err)
		return links, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var link model.Link
		if err := rows.Scan(&link.LinkID, &link.Slug, &link.ShortCode, &link.Destination, &link.CreatedAt, &link.Is_active); err != nil {
			logger.Error("ListLinks: row scan failed: %v", err)
			return links, fmt.Errorf("scan failed: %w", err)
		}
		links = append(links, link)
	}

	if err := rows.Err(); err != nil {
		logger.Error("ListLinks: rows iteration error: %v", err)
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return links, nil
}

func (s *LinkSvc) ResolveLink(ctx context.Context, shortCode string) (string, string, error) {
	var (
		linkID      string
		destination string
	)

	query := `SELECT id::text, destination FROM links WHERE short_code = $1`
	err := s.DB.QueryRowContext(ctx, query, shortCode).Scan(&linkID, &destination)
	if err == sql.ErrNoRows {
		logger.Warn("ResolveLink: short_code not found: %s", shortCode)
		return "", "", ErrLinkNotFound
	} else if err != nil {
		logger.Error("ResolveLink: DB error: %v", err)
		return "", "", fmt.Errorf("resolve failed: %w", err)
	}

	return destination, linkID, nil
}

func (s *LinkSvc) ResolveUserSlug(ctx context.Context, userID, slug string) (model.Link, error) {
	var link model.Link

	query := `
        SELECT id::text, slug, short_code, destination, created_at
        FROM links
        WHERE user_id = $1 AND slug = $2
    `
	err := s.DB.QueryRowContext(ctx, query, userID, slug).Scan(
		&link.LinkID, &link.Slug, &link.ShortCode, &link.Destination, &link.CreatedAt,
	)
	if err == sql.ErrNoRows {
		logger.Warn("ResolveUserSlug: slug not found for userID=%s: %s", userID, slug)
		return model.Link{}, ErrLinkNotFound
	} else if err != nil {
		logger.Error("ResolveUserSlug: DB error: %v", err)
		return model.Link{}, fmt.Errorf("resolve slug failed: %w", err)
	}

	return link, nil
}

func (s *LinkSvc) TrackClick(ctx context.Context, shortCode, ip, referrer, userAgent string) error {
	query := `
		INSERT INTO clicks (id, link_id, ip, referrer, user_agent, created_at)
		SELECT gen_random_uuid(), id, $2, $3, $4, now()
		FROM links
		WHERE short_code = $1
	`
	res, err := s.DB.ExecContext(ctx, query, shortCode, ip, referrer, userAgent)
	if err != nil {
		logger.Error("TrackClick: failed to insert click: %v", err)
		return fmt.Errorf("track click failed: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		logger.Warn("TrackClick: no link found for short_code=%s", shortCode)
		return ErrLinkNotFound
	}
	return nil
}

func (s *LinkSvc) GetClickCount(ctx context.Context, shortCode string) (int, error) {
	var count int

	query := `
		SELECT COUNT(*)
		FROM clicks c
		JOIN links l ON c.link_id = l.id
		WHERE l.short_code = $1
	`
	err := s.DB.QueryRowContext(ctx, query, shortCode).Scan(&count)
	if err != nil {
		logger.Error("GetClickCount: query failed for short_code=%s: %v", shortCode, err)
		return 0, fmt.Errorf("get click count failed: %w", err)
	}

	return count, nil
}

func (s *LinkSvc) DeleteLink(ctx context.Context, userID, linkID string) error {
	// Ensure ownership before deletion
	checkQuery := `SELECT 1 FROM links WHERE id = $1 AND user_id = $2`
	var exists int
	if err := s.DB.QueryRowContext(ctx, checkQuery, linkID, userID).Scan(&exists); err != nil {
		if err == sql.ErrNoRows {
			logger.Warn("DeleteLink: link not found or access denied for linkID=%s, userID=%s", linkID, userID)
			return ErrLinkNotFound
		}
		logger.Error("DeleteLink: DB error: %v", err)
		return fmt.Errorf("check link ownership failed: %w", err)
	}

	deleteQuery := `DELETE FROM links WHERE id = $1 AND user_id = $2`
	_, err := s.DB.ExecContext(ctx, deleteQuery, linkID, userID)
	if err != nil {
		logger.Error("DeleteLink: deletion failed for linkID=%s, userID=%s: %v", linkID, userID, err)
		return fmt.Errorf("delete failed: %w", err)
	}

	return nil
}
