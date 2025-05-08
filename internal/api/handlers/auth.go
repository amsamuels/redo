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

		logger.Info("handling login for sub%s", sub)
		// Check if user exists
		userData, err := au.UserService.GetByID(r.Context(), sub)
		if err != nil {
			if err == sql.ErrNoRows {
				logger.Info("user not found, creating new user sub: %s", sub)

				var user model.SignUpRequest
				if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
					logger.Warn("invalid request body during signup sub:%s error:[%s]", sub, err)
					utils.WriteJSONError(w, http.StatusBadRequest, "Invalid request body")
					return
				}

				userData, err = au.UserService.SignUp(r.Context(), sub, user.Email)
				if err != nil {
					logger.Error("failed to fetch user  sub:%s error:[%s]", sub, err)
					utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to fetch created user")
					return
				}

				logger.Info("user successfully created sub:%s", sub)
			} else {
				logger.Error("database error fetching user sub:%s error:[%s]", sub, err)
				utils.WriteJSONError(w, http.StatusInternalServerError, "Database error")
				return
			}
		}

		logger.Info("user authenticated sub:%s, user_id:%s", sub, userData.UserID)
		// ✅ User exists (or just created) — return basic info
		var res model.User
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&res)
	}
}
