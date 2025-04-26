package server

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var slugRegex = regexp.MustCompile(`^[a-zA-Z0-9-_]+$`)

func Wrap(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		h(w, r)
	}
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

func isValidSlug(slug string) bool {
	return slugRegex.MatchString(slug)
}

func isValidURL(destination string) bool {
	_, err := url.ParseRequestURI(destination)
	return err == nil
}

func getPlatform(userAgent string) string {
	if strings.Contains(userAgent, "iPhone") || strings.Contains(userAgent, "iPad") {
		return "iOS"
	} else if strings.Contains(userAgent, "Android") {
		return "Android"
	}
	return "Web"
}
