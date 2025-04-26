package server

import (
	"encoding/json"
	"net/http"

	lru "github.com/hashicorp/golang-lru"
	"redo.ai/internal/api/handlers"
	"redo.ai/internal/service/link"
	"redo.ai/internal/service/user"
)

type HandlerContainer struct {
	AuthHandler *handlers.AuthHandler
	LinkHandler *handlers.LinkHandler
	Cache       *lru.Cache
	//MetricsHandler *handlers.MetricsHandler
}

func NewHandlerContainer(linkSvc link.LinkService, userSvc user.UserService, c *lru.Cache) *HandlerContainer {
	return &HandlerContainer{
		AuthHandler: handlers.NewAuthHandler(userSvc, c),
		LinkHandler: handlers.NewLinkHandler(linkSvc, c),

		//MetricsHandler: handlers.NewMetricsHandler(linkSvc),
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
