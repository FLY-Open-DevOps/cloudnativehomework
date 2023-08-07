package bff

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

type httpheader string

const (
	headerKey httpheader = "headers"
)

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		id := uuid.New()
		requestTime := time.Now()
		s.logger.Info(
			"start",
			"id", id,
			"path", req.URL.Path,
			"client", req.RemoteAddr,
		)
		next.ServeHTTP(res, req)
		s.logger.Info(
			"end",
			"id", id,
			"path", req.URL.Path,
			"client", req.RemoteAddr,
			"duration", time.Since(requestTime),
		)
	})
}
