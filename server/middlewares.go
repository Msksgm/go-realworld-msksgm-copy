package server

import (
	"io"
	"net/http"

	"github.com/gorilla/handlers"
)

func Logger(w io.Writer) func(h http.Handler) http.Handler {
	return (func(h http.Handler) http.Handler {
		return handlers.LoggingHandler(w, h)
	})
}
