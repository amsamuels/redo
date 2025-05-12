package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	lru "github.com/hashicorp/golang-lru"
	"redo.ai/internal/api/middleware"
	"redo.ai/internal/model"
	"redo.ai/internal/service/user"
	"redo.ai/internal/utils"
	"redo.ai/logger"
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

// LoginHandler - Find or create user based on Auth0 sub
func (au *AuthHandler) LoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sub, ok := middleware.SubFromContext(r.Context())
		if !ok {
			logger.Warn("unauthorized request: missing sub")
			utils.WriteJSONError(w, http.StatusUnauthorized, "Unauthorized: missing sub")
			return
		}

		logger.Info("Handling login for sub: %s", sub)

		// 1. Check LRU cache first
		if cachedUser, ok := au.Cache.Get(sub); ok {
			userData := cachedUser.(model.User)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"userid": userData.UserID,
				"role":   userData.Role,
			})
			return
		}

		// 2. If not cached, fetch from DB or create new user
		userData, err := au.UserService.GetByID(r.Context(), sub)
		if err != nil {
			if err == sql.ErrNoRows {
				logger.Info("User not found in DB, creating new user for sub: %s", sub)
				// 3. If the user doesn't exit sign them up
				var user model.SignUpRequest
				if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
					logger.Warn("Invalid request body during signup for sub: %s, error: %v", sub, err)
					utils.WriteJSONError(w, http.StatusBadRequest, "Invalid request body")
					return
				}

				userData, err = au.UserService.SignUp(r.Context(), sub, user.Email)
				if err != nil {
					logger.Error("Failed to create user for sub: %s, error: %v", sub, err)
					utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to create user")
					return
				}
			} else {
				logger.Error("Database error fetching user for sub: %s, error: %v", sub, err)
				utils.WriteJSONError(w, http.StatusInternalServerError, "Database error")
				return
			}
		}

		// 4. Cache the user data
		au.Cache.Add(sub, *userData)
		logger.Info("User authenticated and cached for sub: %s", sub)

		// 5. Return user data
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"userid": userData.UserID,
			"role":   userData.Role,
		})
	}
}
