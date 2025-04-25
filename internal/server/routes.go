package server

func (s *Server) routes() {
	auth := EnsureValidToken()
	withCompany := WithCompany(s.DB)

	s.Mux.Handle("/api/links", auth(withCompany(Wrap(CreateLinkHandler(s.LinkSvc)))))
	s.Mux.Handle("/api/metrics/", auth(withCompany(Wrap(GetMetricsHandler(s.LinkSvc)))))
	s.Mux.HandleFunc("/go/", RedirectHandler(s.LinkSvc)) // public redirect
}
