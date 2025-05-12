package server

import "redo.ai/internal/api/middleware"

func (s *Server) routes() {

	hc := s.HC // Access the HandlerContainer
	// Public routes (no auth)
	s.Mux.HandleFunc("/go/", hc.LinkHandler.RedirectHandler().ServeHTTP)
	s.Mux.HandleFunc("/api/health", s.HealthHandler())

	// User-related
	// User-related routes

	s.Mux.Handle("/api/user", middleware.ValidateJWT("https://api.mybackend.com", "dev-omr1iha4te137r50.us.auth0.com", hc.AuthHandler.LoginHandler()))

	//Link-related (protected by auth)
	s.Mux.Handle("/api/links", middleware.ValidateJWT("https://api.mybackend.com", "dev-omr1iha4te137r50.us.auth0.com", hc.LinkHandler.LinksRouter()))
	// s.Mux.Handle("/api/links/list", auth(withUser(hc.LinkHandler.ListLinksHandler())))
	//s.Mux.Handle("/api/links/", auth(withUser(hc.LinkHandler.GetMetricsHandler())))
}
