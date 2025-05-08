package utils

import (
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"redo.ai/logger"
)

var SlugRegex = regexp.MustCompile(`^[a-zA-Z0-9-_]+$`)

func LoggingWrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		if r.URL.Path == "/api/health" {
			next.ServeHTTP(w, r)
			return
		}

		logger.Info("Request started: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		logger.Info("Request completed: %s %s (%v)", r.Method, r.URL.Path, duration)
	})
}

func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
func WithCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4040")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

		// Respond to preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Helper function to write JSON errors.
func WriteJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(`{"error":"` + message + `"}`))
}
func WriteJSON(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}
func IsValidSlug(slug string) bool {
	return SlugRegex.MatchString(slug)
}

func IsValidURL(destination string) bool {
	_, err := url.ParseRequestURI(destination)
	return err == nil
}
