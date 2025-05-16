package mock

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	lru "github.com/hashicorp/golang-lru"
	"redo.ai/internal/api/handlers"
	"redo.ai/internal/model"
)

type ctxKey string

const UserIDKey ctxKey = "user_id"

var SubContextKey ctxKey = "sub"

// Local test version of userIDKey (because we can't import unexported things)

// Mock LinkService for testing
type mockLinkService struct{}

// DeleteLink implements link.LinkService.
func (m *mockLinkService) DeleteLink(ctx context.Context, userID string, linkID string) error {
	panic("unimplemented")
}

// ResolveUserSlug implements link.LinkService.
func (m *mockLinkService) ResolveUserSlug(ctx context.Context, userID string, slug string) (model.Link, error) {
	panic("unimplemented")
}

type mockUserService struct{}

// GetByID implements user.UserService.
func (m *mockUserService) GetByID(ctx context.Context, auth0Sub string) (*model.User, error) {
	panic("unimplemented")
}

// SignUp implements user.UserService.
func (m *mockUserService) SignUp(context.Context, string, string) (*model.User, error) {
	panic("unimplemented")
}

// UserExists implements user.UserService.
func (m *mockUserService) UserExists(ctx context.Context, userID string) (bool, error) {
	panic("unimplemented")
}

func (m *mockLinkService) CreateLink(ctx context.Context, userID string, req model.CreateLinkRequest) (model.Link, error) {
	return model.Link{}, nil // Always succeed
}

func (m *mockLinkService) ResolveLink(ctx context.Context, slug string) (string, string, error) {
	return "https://example.com", "", nil // Return a fake URL
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

// TestCreateLinkHandler uses a table-driven format to test various scenarios.
func TestCreateLinkHandler(t *testing.T) {
	// Create a mock service and cache.
	mockLinkSvc := &mockLinkService{}
	mockUserSvc := &mockUserService{}
	cache, _ := lru.New(100) // Mock cache
	handler := handlers.NewLinkHandler(mockUserSvc, mockLinkSvc, cache).CreateLinkHandler()

	// Define test cases in a table-driven format.
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

	// Iterate over the test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal the payload into JSON.
			body, err := json.Marshal(tt.payload)
			if err != nil {
				t.Fatalf("failed to marshal payload: %v", err)
			}

			// Create a new HTTP request.
			req := httptest.NewRequest(http.MethodPost, "/api/links", bytes.NewReader(body))

			// Simulate authenticated context if required.
			if tt.authenticated {
				ctx := context.WithValue(req.Context(), UserIDKey, "test-user-id")
				req = req.WithContext(ctx)
			}

			// Create a response recorder to capture the response.
			rec := httptest.NewRecorder()

			// Serve the HTTP request.
			handler.ServeHTTP(rec, req)

			// Assert the expected status code.
			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}
