// handlers/click_handler.go
package handlers

import (
	"net/http"
	"strings"

	"redo.ai/internal/model"
	"redo.ai/internal/service/clicks"
	"redo.ai/internal/service/user"
	"redo.ai/internal/utils"
	"redo.ai/logger"
)

type ClickHandler struct {
	ClickService clicks.ClickService
	UserService  user.UserService
}

func NewClickHandler(cs clicks.ClickService, us user.UserService) *ClickHandler {
	return &ClickHandler{
		ClickService: cs,
		UserService:  us,
	}
}

func (h *ClickHandler) ClicksRouter() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !validateMethod(w, r, http.MethodGet) {
			return
		}
		if _, ok := verifySubFromContext(w, r); !ok {
			utils.WriteJSONError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		userID, ok := extractUserIDFromRequest(w, r)
		if !ok {
			utils.WriteJSONError(w, http.StatusUnauthorized, "Invalid user")
			return
		}
		usr, err := h.UserService.GetByID(r.Context(), userID)
		if err != nil {
			logger.Error("ClicksRouter: failed to fetch user: %v", err)
			utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to verify user")
			return
		}
		view := r.URL.Query().Get("view")
		switch strings.ToLower(view) {
		case "per-day":
			h.handleClicksPerDay(w, r, usr.UserID, usr.Role)
		case "by-country":
			h.handleGroupedClicks(w, r, usr.UserID, usr.Role, "country")
		case "by-device":
			h.handleGroupedClicks(w, r, usr.UserID, usr.Role, "device")
		default:
			utils.WriteJSONError(w, http.StatusNotFound, "Unknown analytics view")
		}
	}
}

func (h *ClickHandler) handleClicksPerDay(w http.ResponseWriter, r *http.Request, userID string, plan string) {
	if plan == "free" {
		utils.WriteJSONError(w, http.StatusForbidden, "Upgrade to access analytics")
		return
	}
	results, err := h.ClickService.ClicksPerDay(r.Context(), userID)
	if err != nil {
		logger.Error("handleClicksPerDay: error: %v", err)
		utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to load daily clicks")
		return
	}
	utils.WriteJSON(w, http.StatusOK, results)
}

func (h *ClickHandler) handleGroupedClicks(w http.ResponseWriter, r *http.Request, userID, plan, groupBy string) {
	if plan == "free" {
		utils.WriteJSONError(w, http.StatusForbidden, "Upgrade to access analytics")
		return
	}
	var (
		results []model.GroupedMetric
		err     error
	)
	switch groupBy {
	case "country":
		results, err = h.ClickService.GetClicksGroupedByCountry(r.Context(), userID)
	case "device":
		results, err = h.ClickService.GetClicksGroupedByDevice(r.Context(), userID)
	default:
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid grouping")
		return
	}
	if err != nil {
		logger.Error("handleGroupedClicks (%s): %v", groupBy, err)
		utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to load analytics")
		return
	}
	utils.WriteJSON(w, http.StatusOK, results)
}
