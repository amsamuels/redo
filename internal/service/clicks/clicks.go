package clicks

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"redo.ai/internal/model"
	"redo.ai/internal/service/user"
	"redo.ai/logger"
)

// ClickService defines the interface for click-related operations.
type ClickService interface {
	ClicksPerDay(ctx context.Context, userID string) ([]model.ClicksByDay, error)
	TrackClick(ctx context.Context, shortCode, ip, referrer, userAgent, deviceType, country, org string, conversion, highValue bool) error
	GetClickCount(ctx context.Context, shortCode string) (int, error)
	GetLinkClicks(ctx context.Context, linkID string) ([]model.Click, error)
	GetRecentClicksByUser(ctx context.Context, userID string, limit int) ([]model.Click, error)
	GetClicksGroupedByDevice(ctx context.Context, userID string) ([]model.GroupedMetric, error)
	GetClicksGroupedByCountry(ctx context.Context, userID string) ([]model.GroupedMetric, error)
}

var ErrLinkNotFound = errors.New("link not found")

type ClickSvc struct {
	DB          *sql.DB
	UserService user.UserService
}

func (s *ClickSvc) ClicksPerDay(ctx context.Context, userID string) ([]model.ClicksByDay, error) {

	query := `
		SELECT
			to_char(c.created_at::date, 'Dy') AS day,
			c.created_at::date as date_key,
			COUNT(*) AS click_count
		FROM clicks c
		JOIN links l ON c.link_id = l.id
		WHERE l.user_id = $1
		  AND c.created_at >= CURRENT_DATE - INTERVAL '6 days'
		GROUP BY date_key
		ORDER BY date_key;
	`

	rows, err := s.DB.QueryContext(ctx, query, userID)
	if err != nil {
		logger.Error("ClicksPerDay: query failed: %v", err)
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var results []model.ClicksByDay = make([]model.ClicksByDay, 0)
	for rows.Next() {
		var point model.ClicksByDay
		var dateKey time.Time
		if err := rows.Scan(&point.DateLabel, &dateKey, &point.Clicks); err != nil {
			logger.Error("ClicksPerDay: scan failed: %v", err)
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		results = append(results, point)
	}

	if err := rows.Err(); err != nil {
		logger.Error("ClicksPerDay: row iteration failed: %s", err)
		return nil, fmt.Errorf("row iteration failed: %w", err)
	}

	return results, nil
}

func (s *ClickSvc) GetClickCount(ctx context.Context, shortCode string) (int, error) {
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

func (s *ClickSvc) TrackClick(ctx context.Context, shortCode, ip, referrer, userAgent, deviceType, country, org string, conversion, highValue bool) error {
	query := `
		INSERT INTO clicks (link_id, ip, referrer, user_agent, device_type, country, org_name, conversion, is_high_value)
		SELECT id, $2, $3, $4, $5, $6, $7, $8, $9
		FROM links WHERE short_code = $1
	`
	_, err := s.DB.ExecContext(ctx, query, shortCode, ip, referrer, userAgent, deviceType, country, org, conversion, highValue)
	if err != nil {
		logger.Error("TrackClick: failed to insert click for short_code=%s: %v", shortCode, err)
		return fmt.Errorf("track click failed: %w", err)
	}
	return nil
}

func (s *ClickSvc) GetLinkClicks(ctx context.Context, linkID string) ([]model.Click, error) {
	query := `
		SELECT id::text, link_id::text, ip, referrer, user_agent, device_type, country, conversion, is_high_value, created_at
		FROM clicks
		WHERE link_id = $1
		ORDER BY created_at DESC;
	`
	rows, err := s.DB.QueryContext(ctx, query, linkID)
	if err != nil {
		logger.Error("GetLinkClicks: query failed: %v", err)
		return nil, fmt.Errorf("GetLinkClicks: query failed: %w", err)
	}
	defer rows.Close()
	var clicks []model.Click = make([]model.Click, 0)
	for rows.Next() {
		var c model.Click
		if err := rows.Scan(&c.ID, &c.LinkID, &c.IP, &c.Referrer, &c.UserAgent, &c.DeviceType, &c.Country, &c.Conversion, &c.IsHighValue, &c.CreatedAt); err != nil {
			logger.Error("GetLinkClicks: scan failed: %v", err)
			return nil, fmt.Errorf("GetLinkClicks: scan failed: %w", err)
		}
		clicks = append(clicks, c)
	}
	return clicks, nil
}

func (s *ClickSvc) GetRecentClicksByUser(ctx context.Context, userID string, limit int) ([]model.Click, error) {
	query := `
		SELECT c.id::text, c.link_id::text, c.ip, c.referrer, c.user_agent, c.device_type, c.country, c.conversion, c.is_high_value, c.created_at
		FROM clicks c
		JOIN links l ON c.link_id = l.id
		WHERE l.user_id = $1
		ORDER BY c.created_at DESC
		LIMIT $2;
	`
	rows, err := s.DB.QueryContext(ctx, query, userID, limit)
	if err != nil {
		logger.Error("GetRecentClicksByUser: query failed: %v", err)
		return nil, fmt.Errorf("GetRecentClicksByUser: query failed: %w", err)
	}
	defer rows.Close()

	var clicks []model.Click = make([]model.Click, 0)
	for rows.Next() {
		var c model.Click
		if err := rows.Scan(&c.ID, &c.LinkID, &c.IP, &c.Referrer, &c.UserAgent, &c.DeviceType, &c.Country, &c.Conversion, &c.IsHighValue, &c.CreatedAt); err != nil {
			logger.Error("GetRecentClicksByUser: scan failed: %v", err)
			return nil, fmt.Errorf("GetRecentClicksByUser: scan failed: %w", err)
		}
		clicks = append(clicks, c)
	}
	return clicks, nil
}

func (s *ClickSvc) GetClicksGroupedByDevice(ctx context.Context, userID string) ([]model.GroupedMetric, error) {
	query := `
		SELECT COALESCE(device_type, 'unknown') AS label, COUNT(*)
		FROM clicks c
		JOIN links l ON c.link_id = l.id
		WHERE l.user_id = $1
		GROUP BY device_type;
	`
	rows, err := s.DB.QueryContext(ctx, query, userID)
	if err != nil {
		logger.Error("GetClicksGroupedByDevice: query failed: %v", err)
		return nil, fmt.Errorf("GetClicksGroupedByDevice: query failed: %w", err)
	}
	defer rows.Close()

	var results []model.GroupedMetric = make([]model.GroupedMetric, 0)
	for rows.Next() {
		var g model.GroupedMetric
		if err := rows.Scan(&g.Label, &g.Count); err != nil {
			logger.Error("GetClicksGroupedByDevice: scan failed: %v", err)
			return nil, fmt.Errorf("GetClicksGroupedByDevice: scan failed: %w", err)
		}
		results = append(results, g)
	}
	return results, nil
}

func (s *ClickSvc) GetClicksGroupedByCountry(ctx context.Context, userID string) ([]model.GroupedMetric, error) {
	query := `
		SELECT COALESCE(country, 'unknown') AS label, COUNT(*)
		FROM clicks c
		JOIN links l ON c.link_id = l.id
		WHERE l.user_id = $1
		GROUP BY country;
	`
	rows, err := s.DB.QueryContext(ctx, query, userID)
	if err != nil {
		logger.Error("GetClicksGroupedByCountry: query failed: %v", err)
		return nil, fmt.Errorf("GetClicksGroupedByCountry: query failed: %w", err)
	}
	defer rows.Close()

	var results []model.GroupedMetric = make([]model.GroupedMetric, 0)
	for rows.Next() {
		var g model.GroupedMetric
		if err := rows.Scan(&g.Label, &g.Count); err != nil {
			logger.Error("GetClicksGroupedByCountry: scan failed: %v", err)
			return nil, fmt.Errorf("GetClicksGroupedByCountry: scan failed: %w", err)
		}
		results = append(results, g)
	}
	return results, nil
}
