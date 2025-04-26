package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	lru "github.com/hashicorp/golang-lru"
	"redo.ai/internal/api/middleware"
	"redo.ai/internal/model"
	"redo.ai/internal/service/link"
	"redo.ai/internal/utils"
)

type LinkHandler struct {
	LinkService link.LinkService
	Cache       *lru.Cache
}

func NewLinkHandler(linkService link.LinkService, c *lru.Cache) *LinkHandler {
	return &LinkHandler{
		LinkService: linkService,
		Cache:       c,
	}
}

func (lh *LinkHandler) CreateLinkHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			utils.WriteJSONError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
			return
		}

		var req model.CreateLinkRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteJSONError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}

		// Validate inputs
		if !utils.IsValidSlug(req.Slug) {
			utils.WriteJSONError(w, http.StatusBadRequest, "Invalid slug format")
			return
		}
		if !utils.IsValidURL(req.Destination) {
			utils.WriteJSONError(w, http.StatusBadRequest, "Invalid destination URL")
			return
		}

		// Get user ID from context
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			utils.WriteJSONError(w, http.StatusUnauthorized, "Unauthorized")
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

		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			utils.WriteJSONError(w, http.StatusUnauthorized, "Unauthorized")
			return
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
func (lh *LinkHandler) RedirectHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Validate path and slug (existing logic)
		if !strings.HasPrefix(r.URL.Path, "/go/") {
			utils.WriteJSONError(w, http.StatusNotFound, "Invalid path")
			return
		}

		slug := strings.TrimPrefix(r.URL.Path, "/go/")
		if slug == "" || !utils.IsValidSlug(slug) {
			utils.WriteJSONError(w, http.StatusBadRequest, "Invalid slug")
			return
		}

		// Resolve destination URL (e.g., Spotify track URL)
		destURL, ok := lh.Cache.cacheGet(slug)
		if !ok {
			var err error
			destURL, err = lh.LinkService.ResolveLink(r.Context(), slug)
			if err != nil {
				utils.WriteJSONError(w, http.StatusNotFound, "Link not found")
				return
			}
			s.cacheSet(slug, destURL)
		}

		// Detect platform
		userAgent := r.UserAgent()
		platform := utils.GetPlatform(userAgent)

		// Generate deep links or universal links
		var redirectURL string
		switch platform {
		case "iOS":
			redirectURL = fmt.Sprintf("spotify://track/%s", destURL) // Replace with actual track ID
		case "Android":
			redirectURL = fmt.Sprintf("intent://track/%s#Intent;scheme=spotify;package=com.spotify.music;end", destURL)
		default:
			redirectURL = fmt.Sprintf("https://open.spotify.com/track/%s", destURL) // Fallback to web
		}

		// Async click tracking
		go func() {
			_ = lh.LinkService.TrackClick(context.Background(), slug, r.RemoteAddr, r.Referer(), r.UserAgent())
		}()

		// Set headers and redirect
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		http.Redirect(w, r, redirectURL, http.StatusFound)
	}
}
