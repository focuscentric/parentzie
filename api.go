package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func auth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-Api-Key")
		log.Printf("Inside middleware with key %s", key)
		if len(key) == 0 || key != "1234" {
			respond(w, r, http.StatusUnauthorized, nil)
		} else {
			h.ServeHTTP(w, r)
		}
	})
}

func parseBody(body io.ReadCloser, v interface{}) error {
	defer body.Close()

	decoder := json.NewDecoder(body)
	return decoder.Decode(v)
}

func respond(w http.ResponseWriter, r *http.Request, status int, data interface{}) error {
	if err, ok := data.(error); ok {
		data = struct {
			Err string `json:"error"`
		}{err.Error()}
	}
	js, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}
