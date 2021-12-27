package main

import (
	"log"
	"os"

	"github.com/msksgm/go-realworld-msksgm-copy/server"
)

type config struct {
	port  string
	dbURI string
}

func main() {
	cfg := envConfig()

	srv := server.NewServer()
	log.Fatal(srv.Run(cfg.port))
}

func envConfig() config {
	port, ok := os.LookupEnv("PORT")

	if !ok {
		panic("PORT not provided")
	}

	dbURI, ok := os.LookupEnv("POSTGRESQL_URL")

	if !ok {
		panic("POSTGRESQL_URL not provided")
	}

	return config{port: port, dbURI: dbURI}
}
