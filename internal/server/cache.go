package server

func (s *Server) cacheSet(slug, url string) {
	s.cache.Add(slug, url)
}

func (s *Server) cacheGet(slug string) (string, bool) {
	val, ok := s.cache.Get(slug)
	if !ok {
		return "", false
	}
	return val.(string), true
}
