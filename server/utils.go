package server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type M map[string]interface{}

func writeJSON(w http.ResponseWriter, code int, data interface{}) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		serverError(w, err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(jsonBytes)

	if err != nil {
		log.Println(err)
	}
}

func readJSON(body io.Reader, input interface{}) error {
	return json.NewDecoder(body).Decode(input)
}
