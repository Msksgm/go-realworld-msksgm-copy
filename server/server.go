package server

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/msksgm/go-realworld-msksgm-copy/conduit"
	"github.com/msksgm/go-realworld-msksgm-copy/postgres"
)

type Server struct {
	server      *http.Server
	router      *mux.Router
	userService conduit.UserService
}

func NewServer(db *postgres.DB) *Server {
	s := Server{
		server: &http.Server{
			WriteTimeout: 5 * time.Second,
			ReadTimeout:  5 * time.Second,
			IdleTimeout:  5 * time.Second,
		},
		router: mux.NewRouter().StrictSlash(true),
	}

	s.routes()

	s.userService = postgres.NewUserService(db)
	s.server.Handler = s.router

	return &s
}

func (s *Server) Run(port string) error {
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}
	s.server.Addr = port
	log.Printf("server starting on %s", port)
	return s.server.ListenAndServe()
}

func healthCheck() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		resp := M{
			"status":  "available",
			"message": "healthy",
			"data":    M{"hello": "beautiful"},
		}
		writeJSON(rw, http.StatusOK, resp)
	})
}
