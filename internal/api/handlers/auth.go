package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	lru "github.com/hashicorp/golang-lru"
	"redo.ai/internal/api/middleware"
	"redo.ai/internal/model"
	"redo.ai/internal/service/user"
	"redo.ai/internal/utils"
)

type AuthHandler struct {
	UserService user.UserService
	Cache       *lru.Cache
}

func NewAuthHandler(userService user.UserService, cache *lru.Cache) *AuthHandler {
	return &AuthHandler{
		UserService: userService,
		Cache:       cache,
	}
}

// SignUpHandler - Creates a new user (requires valid JWT and business name).

func (au *AuthHandler) SignUpHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("📩 Received signup request")
		log.Println("🔎 Authorization header:", r.Header.Get("Authorization"))
		if r.Method != http.MethodPost {
			utils.WriteJSONError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
			return
		}

		userID, ok := middleware.SubFromContext(r.Context())
		log.Printf("🧪 sub: %s ", userID)
		if !ok {
			utils.WriteJSONError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		var req model.SignUpRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteJSONError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		err := au.UserService.SignUp(r.Context(), userID, req)
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
			utils.WriteJSONError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
			return
		}

		sub, ok := middleware.SubFromContext(r.Context())
		if !ok {
			utils.WriteJSONError(w, http.StatusUnauthorized, "Unauthorized: invalid token")
			return
		}

		user, err := au.UserService.GetByID(r.Context(), sub)
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
