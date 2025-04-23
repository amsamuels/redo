package server

import (
	"net/http"
	"strings"

	"redo.ai/internal/service"
)

func RedirectHandler(svc *service.LinkService) http.HandlerFunc {
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

		url, err := svc.ResolveLink(r.Context(), slug)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		_ = svc.TrackClick(
			r.Context(),
			slug,
			r.RemoteAddr,
			r.Referer(),
			r.UserAgent(),
		)

		http.Redirect(w, r, url, http.StatusFound)

	}
}
