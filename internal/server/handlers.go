package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"redo.ai/internal/model"
)

func (s *Server) HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
		})
	}
}

func (s *Server) CreateLinkHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSONError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
			return
		}

		var req model.CreateLinkRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}

		// Validate inputs
		if !isValidSlug(req.Slug) {
			writeJSONError(w, http.StatusBadRequest, "Invalid slug format")
			return
		}
		if !isValidURL(req.Destination) {
			writeJSONError(w, http.StatusBadRequest, "Invalid destination URL")
			return
		}

		// Get user ID from context
		userID, ok := UserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		// Save link
		if err := s.LinkSvc.CreateLink(r.Context(), userID, req); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "Failed to create link")
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

func (s *Server) ListLinksHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSONError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
			return
		}

		userID, ok := UserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		links, err := s.LinkSvc.ListLinks(r.Context(), userID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "Failed to fetch links")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(links)
	}
}

func (s *Server) RedirectHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Validate path and slug (existing logic)
		if !strings.HasPrefix(r.URL.Path, "/go/") {
			writeJSONError(w, http.StatusNotFound, "Invalid path")
			return
		}

		slug := strings.TrimPrefix(r.URL.Path, "/go/")
		if slug == "" || !isValidSlug(slug) {
			writeJSONError(w, http.StatusBadRequest, "Invalid slug")
			return
		}

		// Resolve destination URL (e.g., Spotify track URL)
		destURL, ok := s.cacheGet(slug)
		if !ok {
			var err error
			destURL, err = s.LinkSvc.ResolveLink(r.Context(), slug)
			if err != nil {
				writeJSONError(w, http.StatusNotFound, "Link not found")
				return
			}
			s.cacheSet(slug, destURL)
		}

		// Detect platform
		userAgent := r.UserAgent()
		platform := getPlatform(userAgent)

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
			_ = s.LinkSvc.TrackClick(context.Background(), slug, r.RemoteAddr, r.Referer(), r.UserAgent())
		}()

		// Set headers and redirect
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		http.Redirect(w, r, redirectURL, http.StatusFound)
	}
}

func (s *Server) GetMetricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		slug := strings.TrimPrefix(r.URL.Path, "/api/metrics/")
		if slug == "" {
			http.Error(w, "Missing slug", http.StatusBadRequest)
			return
		}

		metrics, err := s.LinkSvc.GetClickCount(r.Context(), slug)
		if err != nil {
			http.Error(w, "Error fetching metrics", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]int{"clicks": metrics})
	}
}
