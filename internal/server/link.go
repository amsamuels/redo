package server

import (
	"encoding/json"
	"net/http"

	"redo.ai/internal/model"
	"redo.ai/internal/service"
)

func CreateLinkHandler(svc *service.LinkService) http.HandlerFunc {
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

		// TEMP: fake auth via header
		companyID := r.Header.Get("X-Company-ID")
		if companyID == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if err := svc.CreateLink(r.Context(), companyID, req); err != nil {
			http.Error(w, "Failed to create link", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}
