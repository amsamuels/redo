package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"redo.ai/internal/model"
	"redo.ai/internal/server" // <--- ADD THIS
)

// Local test version of userIDKey (because we can't import unexported things)
type ctxKey string

const userIDKey ctxKey = "user_id"

// Mock LinkService for testing
type mockLinkService struct{}

func (m *mockLinkService) CreateLink(ctx context.Context, userID string, req model.CreateLinkRequest) error {
	return nil // Always succeed
}

func (m *mockLinkService) ResolveLink(ctx context.Context, slug string) (string, error) {
	return "https://example.com", nil // Return a fake URL
}

func (m *mockLinkService) TrackClick(ctx context.Context, slug, ip, referrer, userAgent string) error {
	return nil // Do nothing
}

func (m *mockLinkService) GetClickCount(ctx context.Context, slug string) (int, error) {
	return 42, nil // Return dummy click count
}

func (m *mockLinkService) ListLinks(ctx context.Context, userID string) ([]model.Link, error) {
	return []model.Link{}, nil
}

// Test CreateLinkHandler with good and bad inputs
func TestCreateLinkHandler(t *testing.T) {

	server := &server.Server{
		LinkSvc: &mockLinkService{},
	}

	handler := server.CreateLinkHandler()

	tests := []struct {
		name           string
		payload        model.CreateLinkRequest
		authenticated  bool
		expectedStatus int
	}{
		{
			name: "Valid link creation",
			payload: model.CreateLinkRequest{
				Slug:        "my-awesome-link",
				Destination: "https://example.com",
			},
			authenticated:  true,
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Invalid slug (bad chars)",
			payload: model.CreateLinkRequest{
				Slug:        "bad slug!!",
				Destination: "https://example.com",
			},
			authenticated:  true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid destination URL",
			payload: model.CreateLinkRequest{
				Slug:        "validslug",
				Destination: "not-a-real-url",
			},
			authenticated:  true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Unauthenticated request",
			payload: model.CreateLinkRequest{
				Slug:        "validslug",
				Destination: "https://example.com",
			},
			authenticated:  false,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/api/links", bytes.NewReader(body))

			// Simulate authenticated context
			if tt.authenticated {
				ctx := context.WithValue(req.Context(), userIDKey, "test-user-id")
				req = req.WithContext(ctx)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("%s: expected status %d, got %d", tt.name, tt.expectedStatus, rec.Code)
			}
		})
	}
}
