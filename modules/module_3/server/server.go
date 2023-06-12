package server

import (
	"net/http"
)

const VERSION_ENV = "VERSION"

func Serve(addr string) error {
	http.HandleFunc("/healthz", logger(healthz).Wrap().ServeHTTP)
	return http.ListenAndServe(addr, nil)
}
