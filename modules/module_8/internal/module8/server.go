package module8

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/exp/slog"
)

type Server struct {
	fibo   *FiboCaculator
	port   int
	logger *slog.Logger
	mux    *http.ServeMux
	srv    *http.Server
}

func NewServer(cfg *Config) *Server {
	s := new(Server)
	s.mux = http.NewServeMux()
	s.srv = &http.Server{}
	s.port = cfg.Port
	opts := &slog.HandlerOptions{}
	if cfg.Env == "DEV" {
		opts.Level = slog.LevelDebug
	} else {
		opts.Level = slog.LevelInfo
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	s.logger = logger
	s.fibo = NewFiboCaculator(cfg.MaxSeq, cfg.CacheResult, logger)
	s.register(strings.ToLower(cfg.Env))
	return s
}

func (s *Server) Run() error {
	addr := fmt.Sprintf("0.0.0.0:%d", s.port)
	s.logger.Info("server start", "addr", addr)
	s.srv.Addr = addr
	s.srv.Handler = s.mux
	return s.srv.ListenAndServe()
}

func (s *Server) register(env string) {
	if len(env) == 0 {
		env = "prod"
	}
	s.mux.HandleFunc(fmt.Sprintf("/%s/healthz", env), s.logging(s.healthz).Wrap().ServeHTTP)
	s.mux.HandleFunc(fmt.Sprintf("/%s/fibo", env), s.logging(s.fiboHandler).Wrap().ServeHTTP)
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("graceful stopping server")
	return s.srv.Shutdown(ctx)
}

func (s *Server) healthz(w *Response, r *http.Request) {
	if r.Method != http.MethodGet {
		w.SetStatusCode(http.StatusMethodNotAllowed)
		return
	}
	w.SetStatusCode(http.StatusOK)
}

func (s *Server) fiboHandler(w *Response, r *http.Request) {
	if r.Method != http.MethodGet {
		w.SetStatusCode(http.StatusMethodNotAllowed)
		return
	}
	n, err := strconv.Atoi(r.URL.Query().Get("n"))
	if err != nil {
		s.logger.Error("fibo caculate failed", "errmsg", err.Error())
		w.SetStatusCode(http.StatusBadRequest)
		w.Write([]byte("invalid integer"))
		return
	}
	f, err := s.fibo.Caculate(n)
	if err != nil {
		s.logger.Error("fibo caculate failed", "errmsg", err.Error())
		w.SetStatusCode(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.SetContentType("application/json")
	w.Write([]byte(fmt.Sprintf("{\"result\": %d}", f)))
}

func (s *Server) logging(next Handler) Handler {
	return func(w *Response, r *http.Request) {
		requestID := uuid.New()
		requestTime := time.Now()
		s.logger.Info(
			"start",
			"id", requestID,
			"path", r.URL.Path,
			"client", r.RemoteAddr,
		)
		next(w, r)
		s.logger.Info(
			"end",
			"id", requestID,
			"path", r.URL.Path,
			"client", r.RemoteAddr,
			"status code", w.StatusCode(),
			"duration", time.Since(requestTime),
		)
	}
}
