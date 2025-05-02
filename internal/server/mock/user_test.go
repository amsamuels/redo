package mock

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"redo.ai/internal/api/handlers"
	"redo.ai/internal/api/middleware"
	"redo.ai/internal/model"
)

// MockUserService implements the UserService interface.
type MockUserService struct {
	ShouldConflict bool
	ShouldError    bool
}

func (m *MockUserService) SignUp(context.Context, string, model.SignUpRequest) error {
	if m.ShouldConflict {
		return sql.ErrNoRows // Simulate "already exists" conflict
	}
	if m.ShouldError {
		return errors.New("some db error") // Simulate server-side db error
	}
	return nil
}

func (m *MockUserService) GetByID(ctx context.Context, email string) (*model.User, error) {
	return &model.User{
		ID:           "mock-id",
		Email:        email,
		Name:         "Mock User",
		BusinessName: "Mock Business",
		CreatedAt:    "2023-01-01T00:00:00Z",
	}, nil
}

func newAuthenticatedRequest(method, url string, body []byte, sub string) *http.Request {
	req := httptest.NewRequest(method, url, bytes.NewReader(body))

	// If sub is non-empty, inject it into context
	if sub != "" {
		ctx := context.WithValue(req.Context(), middleware.SubContextKey, sub)
		req = req.WithContext(ctx)
	}

	return req
}

func TestSignUpHandler(t *testing.T) {
	tests := []struct {
		name           string
		mockService    *MockUserService
		payload        model.SignUpRequest
		authenticated  bool
		expectedStatus int
	}{
		{
			name:        "Valid user creation",
			mockService: &MockUserService{},
			payload: model.SignUpRequest{
				Name:         "John Doe",
				BusinessName: "Example Corp",
			},
			authenticated:  true,
			expectedStatus: http.StatusCreated,
		},
		{
			name:        "Duplicate user creation",
			mockService: &MockUserService{ShouldConflict: true},
			payload: model.SignUpRequest{
				Name:         "Jane Doe",
				BusinessName: "Another Corp",
			},
			authenticated:  true,
			expectedStatus: http.StatusConflict,
		},
		{
			name:        "Server error",
			mockService: &MockUserService{ShouldError: true},
			payload: model.SignUpRequest{
				Name:         "Jake Error",
				BusinessName: "Error Corp",
			},
			authenticated:  true,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:        "Unauthenticated request",
			mockService: &MockUserService{},
			payload: model.SignUpRequest{
				Name:         "No Auth",
				BusinessName: "No Auth Corp",
			},
			authenticated:  false,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			handler := handlers.NewAuthHandler(tt.mockService, nil).SignUpHandler()
			body, _ := json.Marshal(tt.payload)

			sub := ""
			if tt.authenticated {
				sub = "user@example.com"
			}
			req := newAuthenticatedRequest(http.MethodPost, "/api/users/signup", body, sub)

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("%s: expected status %d, got %d", tt.name, tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestLogInHandler(t *testing.T) {

	tests := []struct {
		name           string
		mockService    *MockUserService
		authenticated  bool
		subEmail       string
		expectedStatus int
	}{
		{
			name:           "Valid user login",
			mockService:    &MockUserService{},
			authenticated:  true,
			subEmail:       "user@example.com",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Unauthenticated request (no sub)",
			mockService:    &MockUserService{},
			authenticated:  false,
			subEmail:       "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := handlers.NewAuthHandler(tt.mockService, nil).LoginHandler()

			req := newAuthenticatedRequest(http.MethodPost, "/api/users/login", nil, tt.subEmail)

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != tt.expectedStatus {
				t.Errorf("%s: expected status %d, got %d", tt.name, tt.expectedStatus, rec.Code)
			}
		})
	}
}
