package server

import (
	"database/sql"
	"net/http"

	lru "github.com/hashicorp/golang-lru"
	"redo.ai/internal/service"
)

type Server struct {
	DB         *sql.DB
	LinkSvc    LinkService
	UserSvc    UserService
	cache      *lru.Cache
	Mux        *http.ServeMux
	HttpServer *http.Server
}

func New(db *sql.DB) *Server {
	linkSvc := &service.LinkService{DB: db}
	userSvc := &service.UserService{DB: db}

	mux := http.NewServeMux()

	c, _ := lru.New(10000) // cache up to 10,000 links

	srv := &Server{
		DB:      db,
		LinkSvc: linkSvc, // the concrete implementations
		UserSvc: userSvc,
		Mux:     mux,
		cache:   c,
	}

	srv.routes()

	return srv
}

func (s *Server) Start(addr string) error {
	s.HttpServer = &http.Server{
		Addr:    addr,
		Handler: s.Mux,
	}

	return s.HttpServer.ListenAndServe()
}
