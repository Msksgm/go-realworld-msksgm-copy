package main

import (
	"fmt"
	"os"
)

type config struct {
	port  string
	dbURI string
}

func main() {
	cfg := envConfig()
	fmt.Println(cfg)
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
