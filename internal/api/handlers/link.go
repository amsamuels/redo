package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	lru "github.com/hashicorp/golang-lru"
	"redo.ai/internal/api/middleware"
	"redo.ai/internal/model"
	"redo.ai/internal/pkg/platform"
	"redo.ai/internal/service/link"
	"redo.ai/internal/service/user"
	"redo.ai/internal/utils"
	"redo.ai/logger"
)

type LinkHandler struct {
	LinkService link.LinkService
	UserService user.UserService
	Platform    platform.PlatformDetector
	Cache       *lru.Cache
}

func NewLinkHandler(userService user.UserService, linkService link.LinkService, cache *lru.Cache) *LinkHandler {
	return &LinkHandler{
		LinkService: linkService,
		UserService: userService,
		Cache:       cache,
	}
}

func (lh *LinkHandler) LinksRouter() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			lh.CreateLinkHandler().ServeHTTP(w, r)
		case http.MethodGet:
			lh.ListLinksHandler().ServeHTTP(w, r)
		default:
			utils.WriteJSONError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		}
	}
}

func (lh *LinkHandler) CreateLinkHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			utils.WriteJSONError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
			return
		}

		// get sub from context
		_, ok := middleware.SubFromContext(r.Context())
		if !ok {
			utils.WriteJSONError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		logger.Info("recived create link request")

		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			utils.WriteJSONError(w, http.StatusBadRequest, "Missing X-User-ID")
			return
		}

		if _, err := uuid.Parse(userID); err != nil {
			utils.WriteJSONError(w, http.StatusBadRequest, "Invalid X-User-ID format")
			return
		}

		// Check if user exists (if not enforced by DB foreign key)
		if exists, err := lh.UserService.UserExists(r.Context(), userID); err != nil || !exists {
			utils.WriteJSONError(w, http.StatusBadRequest, "UserID does not exist")
		}

		var req model.CreateLinkRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteJSONError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}

		logger.Info("processing slug:%s & destination:%s", req.Slug, req.Destination)

		// Validate inputs
		if !utils.IsValidSlug(req.Slug) || !utils.IsValidURL(req.Destination) {
			utils.WriteJSONError(w, http.StatusBadRequest, "Invalid format")
			return
		}

		// Save link
		if err := lh.LinkService.CreateLink(r.Context(), userID, req); err != nil {
			utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to create link")
			return
		}

		// Build the full masked URL
		maskedURL := fmt.Sprintf("https://%s/go/%s", r.Host, req.Slug)

		// Respond with the masked link
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location", maskedURL)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Link created successfully",
			"link":    maskedURL,
		})
	}
}

func (lh *LinkHandler) ListLinksHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			utils.WriteJSONError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
			return
		}

		_, ok := middleware.SubFromContext(r.Context())
		if !ok {
			utils.WriteJSONError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			utils.WriteJSONError(w, http.StatusBadRequest, "Missing X-User-ID")
			return
		}

		if _, err := uuid.Parse(userID); err != nil {
			utils.WriteJSONError(w, http.StatusBadRequest, "Invalid X-User-ID format")
			return
		}

		// Check if user exists (if not enforced by DB foreign key)
		if exists, err := lh.UserService.UserExists(r.Context(), userID); err != nil || !exists {
			utils.WriteJSONError(w, http.StatusBadRequest, "UserID does not exist")
		}

		links, err := lh.LinkService.ListLinks(r.Context(), userID)
		if err != nil {
			utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to fetch links")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(links)
	}
}
