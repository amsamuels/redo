package handlers

// func (s *Server) GetMetricsHandler() http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		if r.Method != http.MethodGet {
// 			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
// 			return
// 		}

// 		slug := strings.TrimPrefix(r.URL.Path, "/api/metrics/")
// 		if slug == "" {
// 			http.Error(w, "Missing slug", http.StatusBadRequest)
// 			return
// 		}

// 		metrics, err := s.LinkSvc.GetClickCount(r.Context(), slug)
// 		if err != nil {
// 			http.Error(w, "Error fetching metrics", http.StatusInternalServerError)
// 			return
// 		}

// 		json.NewEncoder(w).Encode(map[string]int{"clicks": metrics})
// 	}
// }
