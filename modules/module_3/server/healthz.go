package server

import (
	"net/http"
	"os"
)

func healthz(w *Response, r *http.Request) {
	version := os.Getenv(VERSION_ENV)
	buildResponseHeader(version, w, r)
	w.SetStatusCode(http.StatusOK)
}

func buildResponseHeader(version string, w *Response, r *http.Request) {
	if len(version) > 0 {
		w.SetHeader(VERSION_ENV, version)
	}
	for key, values := range r.Header {
		for _, value := range values {
			w.SetHeader(key, value)
		}
	}
}
