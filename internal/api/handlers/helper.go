package handlers

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"redo.ai/internal/api/middleware"
	"redo.ai/internal/service/user"
	"redo.ai/internal/utils"
	"redo.ai/logger"
)

// validateMethod ensures the request method matches what's expected
func validateMethod(w http.ResponseWriter, r *http.Request, expected string) bool {
	if r.Method != expected {
		logger.Warn("invalid method %s, expected %s", r.Method, expected)
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return false
	}
	return true
}

// extractUserIDFromRequest gets and validates the X-User-ID header
func extractUserIDFromRequest(w http.ResponseWriter, r *http.Request) (string, bool) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		logger.Warn("missing X-User-ID header")
		utils.WriteJSONError(w, http.StatusBadRequest, "Missing X-User-ID")
		return "", false
	}

	if !IsValidUUID(userID) {
		logger.Warn("invalid UUID format for userID: %s", userID)
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid X-User-ID format")
		return "", false
	}
	return userID, true
}

// verifySubFromContext ensures the sub claim is present in the context
func verifySubFromContext(w http.ResponseWriter, r *http.Request) (string, bool) {
	sub, ok := middleware.SubFromContext(r.Context())
	if !ok {
		logger.Warn("missing sub claim in context")
		utils.WriteJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return "", false
	}
	logger.Info("extracted sub from context: %s", sub)
	return sub, true
}

// checkUserExists verifies that the user exists using the UserService
func checkUserExists(ctx context.Context, userService user.UserService, w http.ResponseWriter, userID string) bool {
	exists, err := userService.UserExists(ctx, userID)
	if err != nil {
		logger.Error("error checking if user exists: %v", err)
		utils.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
		return false
	}
	if !exists {
		logger.Warn("userID does not exist: %s", userID)
		utils.WriteJSONError(w, http.StatusBadRequest, "UserID does not exist")
		return false
	}
	return true
}

func IsValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}
