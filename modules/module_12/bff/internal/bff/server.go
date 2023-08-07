package bff

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/exp/slog"
)

var (
	r *rand.Rand
)

func init() {
	r = rand.New(rand.NewSource(time.Now().Unix()))
}

type Server struct {
	bff    *Bff
	logger *slog.Logger
	router *Router
}

func NewServer(cfg *Config, bff *Bff) *Server {
	s := new(Server)
	s.router = New(fmt.Sprintf("0.0.0.0:%d", cfg.Port))
	s.bff = bff
	opts := &slog.HandlerOptions{}
	if cfg.Env == "DEV" {
		opts.Level = slog.LevelDebug
	} else {
		opts.Level = slog.LevelInfo
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	s.logger = logger

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
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	delay()
	n, err := strconv.Atoi(r.URL.Query().Get("n"))
	if err != nil {
		s.logger.Error("fibo caculate failed", "errmsg", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid integer"))
		return
	}
	ctx := context.WithValue(context.Background(), headerKey, r.Header)
	f, err := s.bff.Fibo(ctx, n)
	if err != nil {
		s.logger.Error("fibo caculate failed", "errmsg", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf("{\"result\": %d}", f)))
}

func delay() {
	time.Sleep(time.Millisecond * time.Duration(r.Intn(2000)))
}
