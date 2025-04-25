package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"redo.ai/internal/model"
)

func (s *Server) CreateLinkHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		var req model.CreateLinkRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		companyID, ok := CompanyIDFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if err := s.LinkSvc.CreateLink(r.Context(), companyID, req); err != nil {
			http.Error(w, "Failed to create link", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func (s *Server) RedirectHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/go/") {
			http.NotFound(w, r)
			return
		}

		slug := strings.TrimPrefix(r.URL.Path, "/go/")
		if slug == "" {
			http.Error(w, "Missing slug", http.StatusBadRequest)
			return
		}

		url, err := s.LinkSvc.ResolveLink(r.Context(), slug)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		_ = s.LinkSvc.TrackClick(
			r.Context(),
			slug,
			r.RemoteAddr,
			r.Referer(),
			r.UserAgent(),
		)

		http.Redirect(w, r, url, http.StatusFound)

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
