// internal/server/server.go
package server

import (
	"database/sql"
	"net/http"

	"redo.ai/internal/service"
)

type Server struct {
	DB         *sql.DB
	LinkSvc    *service.LinkService
	Mux        *http.ServeMux
	HttpServer *http.Server
}

func New(db *sql.DB) *Server {
	linkSvc := &service.LinkService{DB: db}
	mux := http.NewServeMux()

	srv := &Server{
		DB:      db,
		LinkSvc: linkSvc,
		Mux:     mux,
	}

	srv.routes() // register all handlers

	return srv
}

func (s *Server) Start(addr string) error {
	s.HttpServer = &http.Server{
		Addr:    addr,
		Handler: s.Mux,
	}

	return s.HttpServer.ListenAndServe()
}
