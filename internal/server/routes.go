package server

func (s *Server) routes() {
	auth := EnsureValidToken()
	withUser := WithUser(s.DB)

	// Public routes (no auth)
	s.Mux.HandleFunc("/go/", s.RedirectHandler())

	// User-related
	s.Mux.Handle("/api/users/signup", s.SignUpHandler())
	s.Mux.Handle("/api/users/login", s.LoginHandler())

	// Link-related (protected by auth)
	s.Mux.Handle("/api/links", auth(withUser(s.CreateLinkHandler())))
	s.Mux.Handle("/api/links/", auth(withUser(s.GetMetricsHandler())))
}
