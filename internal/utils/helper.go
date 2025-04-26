package utils

import (
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var SlugRegex = regexp.MustCompile(`^[a-zA-Z0-9-_]+$`)

func LoggingWrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Skip logging for health checks
		if r.URL.Path == "/api/health" {
			next.ServeHTTP(w, r)
			return
		}

		log.Printf("Request started: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		log.Printf("Request completed: %s %s (%v)", r.Method, r.URL.Path, duration)
	})
}

// Helper function to write JSON errors.
func WriteJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(`{"error":"` + message + `"}`))
}

func IsValidSlug(slug string) bool {
	return SlugRegex.MatchString(slug)
}

func IsValidURL(destination string) bool {
	_, err := url.ParseRequestURI(destination)
	return err == nil
}

func GetPlatform(userAgent string) string {
	if strings.Contains(userAgent, "iPhone") || strings.Contains(userAgent, "iPad") {
		return "iOS"
	} else if strings.Contains(userAgent, "Android") {
		return "Android"
	}
	return "Web"
}
