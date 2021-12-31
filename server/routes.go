package server

import "os"

func (s *Server) routes() {
	s.router.Use(Logger(os.Stdout))
	apiRouter := s.router.PathPrefix("/api/v1").Subrouter()

	noAuth := apiRouter.PathPrefix("").Subrouter()
	{
		noAuth.Handle("/health", healthCheck())
		noAuth.Handle("/users", s.createUser()).Methods("POST")
		noAuth.Handle("/users/login", s.loginUser()).Methods("POST")
	}
}
