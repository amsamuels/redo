package server

import (
	"context"

	"redo.ai/internal/model"
)

type LinkService interface {
	CreateLink(ctx context.Context, userID string, req model.CreateLinkRequest) error
	ResolveLink(ctx context.Context, slug string) (string, error)
	TrackClick(ctx context.Context, slug, ip, referrer, userAgent string) error
	GetClickCount(ctx context.Context, slug string) (int, error)
	ListLinks(ctx context.Context, userID string) ([]model.Link, error)
}

type UserService interface {
	SignUp(ctx context.Context, req model.SignUpRequest) error
	GetByEmail(ctx context.Context, email string) (*model.User, error)
}
