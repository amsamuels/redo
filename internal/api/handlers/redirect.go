package handlers

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
