package server

import (
	"database/sql"
	"net/http"

	lru "github.com/hashicorp/golang-lru"

	"redo.ai/internal/service/link"
	"redo.ai/internal/service/user"
	"redo.ai/internal/utils"
)

type Server struct {
	DB         *sql.DB
	LinkSvc    link.LinkService
	UserSvc    user.UserService
	cache      *lru.Cache
	Mux        *http.ServeMux
	HttpServer *http.Server
	Handler    http.Handler
	HC         *HandlerContainer
}

func New(db *sql.DB) *Server {
	linkSvc := &link.LinkSvc{DB: db}
	userSvc := &user.UserSvc{DB: db}

	mux := http.NewServeMux()

	c, _ := lru.New(10000) // cache up to 10,000 links

	srv := &Server{
		DB:      db,
		LinkSvc: linkSvc,
		UserSvc: userSvc,
		Mux:     mux,
		cache:   c,
	}

	// Initialize handler container with the server instance
	srv.HC = NewHandlerContainer(srv)
	srv.routes()
	// Apply logging middleware globally
	srv.Handler = utils.LoggingWrap(mux)

	return srv
}

func (s *Server) Start(addr string) error {
	s.HttpServer = &http.Server{
		Addr:    addr,
		Handler: s.Handler,
	}
	return s.HttpServer.ListenAndServe()
}
