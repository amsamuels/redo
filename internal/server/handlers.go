package server

import (
	"encoding/json"
	"net/http"

	"redo.ai/internal/api/handlers"
)

type HandlerContainer struct {
	AuthHandler *handlers.AuthHandler
	LinkHandler *handlers.LinkHandler
	//MetricsHandler *handlers.MetricsHandler
}

func NewHandlerContainer(srv *Server) *HandlerContainer {
	return &HandlerContainer{
		AuthHandler: handlers.NewAuthHandler(srv.UserSvc, srv.cache),
		LinkHandler: handlers.NewLinkHandler(srv.UserSvc, srv.LinkSvc, srv.cache),
	}
}

func (s *Server) HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
		})
	}
}
