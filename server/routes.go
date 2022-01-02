package server

import "os"

const MustAuth bool = true

func (s *Server) routes() {
	s.router.Use(Logger(os.Stdout))
	apiRouter := s.router.PathPrefix("/api/v1").Subrouter()

	noAuth := apiRouter.PathPrefix("").Subrouter()
	{
		noAuth.Handle("/health", healthCheck())
		noAuth.Handle("/users", s.createUser()).Methods("POST")
		noAuth.Handle("/users/login", s.loginUser()).Methods("POST")
	}

	authApiRoutes := apiRouter.PathPrefix("").Subrouter()
	authApiRoutes.Use(s.authenticate(MustAuth))
	{
		authApiRoutes.Handle("/user", s.getCurrentUser()).Methods("GET")
		authApiRoutes.Handle("/user", s.updateUser()).Methods("PUT", "PATCH")
		authApiRoutes.Handle("/articles", s.createArticle()).Methods("POST")
		authApiRoutes.Handle("/articles", s.listArticles()).Methods("GET")
	}
}
