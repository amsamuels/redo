package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/golang-jwt/jwt/v5"
	lru "github.com/hashicorp/golang-lru"
	"redo.ai/internal/model"
	"redo.ai/internal/service/user"
	"redo.ai/internal/utils"
)

type AuthHandler struct {
	UserService user.UserService
	Cache       *lru.Cache
}

func NewAuthHandler(userService user.UserService, c *lru.Cache) *AuthHandler {
	return &AuthHandler{
		UserService: userService,
		Cache:       c,
	}
}

// SignUpHandler - Creates a new user (requires valid JWT and business name).
func (au *AuthHandler) SignUpHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		email, ok := SubFromContext(r.Context())
		if !ok {
			utils.WriteJSONError(w, http.StatusBadRequest, "Unauthorized")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req model.SignUpRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteJSONError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		req.Email = email // Force the email from Auth0 token, ignore any email from body.

		err := au.UserService.SignUp(r.Context(), req)
		if err != nil {
			if err == sql.ErrNoRows {
				utils.WriteJSONError(w, http.StatusConflict, "User already exists")
				return
			}
			utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to create user")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"status": "user created"})
	}
}

// LoginHandler - Fetches user details from database based on Auth0 sub (email).
func (au *AuthHandler) LoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		email, ok := SubFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		user, err := au.UserService.GetByEmail(r.Context(), email)
		if err != nil {
			if err == sql.ErrNoRows {
				utils.WriteJSONError(w, http.StatusUnauthorized, "User not found")
				return
			}
			utils.WriteJSONError(w, http.StatusInternalServerError, "Database error")
			return
		}

		resp := model.User{
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
