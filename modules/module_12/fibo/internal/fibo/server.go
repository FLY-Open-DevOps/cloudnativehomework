package fibo

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/exp/slog"
)

type Server struct {
	fibo   *FiboCaculator
	logger *slog.Logger
	router *Router
}

func NewServer(cfg *Config) *Server {
	s := new(Server)
	s.router = New(fmt.Sprintf("0.0.0.0:%d", cfg.Port))
	opts := &slog.HandlerOptions{}
	if cfg.Env == "DEV" {
		opts.Level = slog.LevelDebug
	} else {
		opts.Level = slog.LevelInfo
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	s.logger = logger
	s.fibo = NewFiboCaculator(cfg.MaxSeq, cfg.CacheResult, logger)

	// s.router.RegisterMiddleware()

	s.register(strings.ToLower(cfg.Env))
	return s
}

func (s *Server) Run() error {
	return s.router.Serve()
}

func (s *Server) register(env string) {
	if len(env) == 0 {
		env = "prod"
	}
	prom := NewPromMiddleware(prometheus.ExponentialBuckets(0.1, 1.5, 5))
	s.router.RegisterMiddleware(s.loggingMiddleware)

	s.router.RegisterHandler("/healthz", http.HandlerFunc(s.healthz))
	s.router.RegisterHandler("/fibo", http.HandlerFunc(prom.WrapHandler("/fibo", http.HandlerFunc(s.fiboHandler))))
	s.router.RegisterHandler("/metrics", promhttp.HandlerFor(
		registry,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		}),
	)
}

func (s *Server) Stop(ctx context.Context) error {
	return s.router.Shutdown(ctx)
}

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) fiboHandler(w http.ResponseWriter, r *http.Request) {
	Delay()
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	n, err := strconv.Atoi(r.URL.Query().Get("n"))
	if err != nil {
		s.logger.Error("fibo caculate failed", "errmsg", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid integer"))
		return
	}
	f, err := s.fibo.Caculate(n)
	if err != nil {
		s.logger.Error("fibo caculate failed", "errmsg", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf("{\"result\": %d}", f)))
}
