package server

func (s *Server) routes() {
	s.Mux.HandleFunc("/api/links", Wrap(CreateLinkHandler(s.LinkSvc)))
	s.Mux.HandleFunc("/go/", RedirectHandler(s.LinkSvc))
}
