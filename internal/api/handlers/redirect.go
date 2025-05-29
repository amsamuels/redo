package handlers

import (
	"net/http"
	"strings"

	"redo.ai/internal/service/link"
	"redo.ai/internal/utils"
)

// func (lh *LinkHandler) RedirectHandler() http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		// Validate path and slug (existing logic)
// 		if !strings.HasPrefix(r.URL.Path, "/go/") {
// 			utils.WriteJSONError(w, http.StatusNotFound, "Invalid path")
// 			return
// 		}

// 		slug := strings.TrimPrefix(r.URL.Path, "/go/")
// 		if slug == "" || !utils.IsValidSlug(slug) {
// 			utils.WriteJSONError(w, http.StatusBadRequest, "Invalid slug")
// 			return
// 		}

// 		destURL, ok := lh.Cache.Get(slug)
// 		if !ok {
// 			var err error
// 			destURL, _, err = lh.LinkService.ResolveLink(r.Context(), slug)
// 			if err != nil {
// 				utils.WriteJSONError(w, http.StatusNotFound, "Link not found")
// 				return
// 			}
// 			lh.Cache.Add(slug, destURL) // Use Add to set the value in the cache
// 		}

// 		// Detect platform + service
// 		userAgent := r.UserAgent()
// 		platform := lh.Platform.DetectOs(userAgent)
// 		service := lh.Platform.GetService(destURL.(string))

// 		// Generate best deep link
// 		redirectURL := lh.Platform.GenerateDeepLink(platform, service, destURL.(string))

// 		// Async click tracking
// 		go func() {
// 			_ = lh.LinkService.TrackClick(context.Background(), slug, r.RemoteAddr, r.Referer(), r.UserAgent())
// 		}()

// 		// Set headers and redirect
// 		w.Header().Set("Cache-Control", "no-store")
// 		w.Header().Set("X-Content-Type-Options", "nosniff")
// 		http.Redirect(w, r, redirectURL, http.StatusFound)
// 	}
// }

func (lh *LinkHandler) RedirectHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortCode := strings.TrimPrefix(r.URL.Path, "/r/")
		if shortCode == "" {
			utils.WriteJSONError(w, http.StatusBadRequest, "Missing short code")
			return
		}
		go func() {
			ip := r.RemoteAddr
			ref := r.Referer()
			ua := r.UserAgent()
			_ = lh.LinkService.TrackClick(r.Context(), shortCode, ip, ref, ua)
		}()
		destination, _, err := lh.LinkService.ResolveLink(r.Context(), shortCode)
		if err == link.ErrLinkNotFound {
			utils.WriteJSONError(w, http.StatusNotFound, "Link not found")
			return
		} else if err != nil {
			utils.WriteJSONError(w, http.StatusInternalServerError, "Could not resolve link")
			return
		}
		http.Redirect(w, r, destination, http.StatusFound)
	}
}
