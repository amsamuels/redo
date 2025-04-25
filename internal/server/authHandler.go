package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/golang-jwt/jwt/v5"
	"redo.ai/internal/model"
)

// SignUpHandler - creates a new user with provided business name and associates to Auth0 sub (email as identifier).
type SignUpRequest struct {
	Email        string `json:"email"`
	Name         string `json:"name"`
	BusinessName string `json:"business_name"`
}

type LoginRequest struct {
	Email string `json:"email"`
}

type UserResponse struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	BusinessName string    `json:"business_name"`
	CreatedAt    time.Time `json:"created_at"`
}

func (s *Server) SignUpHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		var req model.SignUpRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		err := s.UserSvc.SignUp(r.Context(), req)
		if err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"status": "user created"})
	}
}

// LoginHandler - fetches user info to confirm they exist, otherwise returns unauthorized.
func (s *Server) LoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		var req model.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		user, err := s.UserSvc.Login(r.Context(), req)
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		resp := UserResponse{
			ID:           user.ID,
			Email:        user.Email,
			Name:         user.Name,
			BusinessName: user.BusinessName,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// --- Helper ---

func SubFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(jwtmiddleware.ContextKey{}).(*jwt.Token)
	if !ok || token == nil {
		return "", false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", false
	}

	sub, ok := claims["sub"].(string)
	return sub, ok
}
