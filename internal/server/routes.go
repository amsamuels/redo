package server

import (
	"net/http"

	"redo.ai/internal/server/api/middleware"
)

func (s *Server) routes() {
	auth := middleware.EnsureValidToken()
	withUser := middleware.WithUser(s.DB)

	hc := s.HC // Access the HandlerContainer
	// Public routes (no auth)

	// s.Mux.HandleFunc("/go/", hc.LinkHandler.Redirect)
	// s.Mux.HandleFunc("/api/health", s.HealthHandler())

	// User-related
	// User-related routes
	s.Mux.Handle("/api/users/signup", http.HandlerFunc(hc.AuthHandler.SignUpHandler()))
	s.Mux.Handle("/api/users/login", http.HandlerFunc(hc.AuthHandler.LoginHandler()))

	// Link-related (protected by auth)
	// s.Mux.Handle("/api/links", auth(withUser(s.CreateLinkHandler())))
	// s.Mux.Handle("/api/links/", auth(withUser(s.GetMetricsHandler())))
	// s.Mux.Handle("/api/links", auth(withUser(s.ListLinksHandler())))
}
