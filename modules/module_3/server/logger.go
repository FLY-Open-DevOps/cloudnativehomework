package server

import (
	"log"
	"net/http"
	"os"
)

var logWriter log.Logger

func init() {
	logWriter = *log.New(os.Stdout, "", 0)
	logWriter.SetPrefix("HTTP SERVER: \n")
}

func logger(next Handler) Handler {
	return func(w *Response, r *http.Request) {
		next(w, r)
		logWriter.Printf("\tPath      : %s\n\tClientIP  : %s\n\tStatusCode: %d", r.URL.Path, r.RemoteAddr, w.StatusCode())
	}
}
