package server

import "redo.ai/internal/api/middleware"

func (s *Server) routes() {
	auth := middleware.EnsureValidToken()
	withUser := middleware.WithUser(s.DB)

	hc := s.HC // Access the HandlerContainer
	// Public routes (no auth)

	s.Mux.HandleFunc("/go/", hc.LinkHandler.RedirectHandler().ServeHTTP)
	s.Mux.HandleFunc("/api/health", s.HealthHandler())

	// User-related
	// User-related routes
	s.Mux.Handle("/api/users/signup", hc.AuthHandler.SignUpHandler())
	s.Mux.Handle("/api/users/login", hc.AuthHandler.LoginHandler())

	//Link-related (protected by auth)
	s.Mux.Handle("/api/links", auth(withUser(hc.LinkHandler.CreateLinkHandler())))
	s.Mux.Handle("/api/links", auth(withUser(hc.LinkHandler.ListLinksHandler())))
	//s.Mux.Handle("/api/links/", auth(withUser(hc.LinkHandler.GetMetricsHandler())))
}
