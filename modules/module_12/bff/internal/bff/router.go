package bff

import (
	"context"
	"log"
	"net/http"
)

type Middleware func(http.Handler) http.Handler

type Router struct {
	addr        string
	middlewares []Middleware
	srv         *http.Server
	mux         *http.ServeMux
}

func New(addr string) *Router {
	if len(addr) == 0 {
		addr = "0.0.0.0:80"
	}
	srv := &http.Server{}
	srv.Addr = addr
	return &Router{
		addr:        addr,
		middlewares: make([]Middleware, 0),
		srv:         srv,
		mux:         http.NewServeMux(),
	}
}

func (r *Router) RegisterMiddleware(m Middleware) *Router {
	r.middlewares = append(r.middlewares, m)
	return r
}

func (r *Router) RegisterHandler(path string, h http.Handler, middlewares ...Middleware) *Router {
	var mergeHandler http.Handler = h
	for i := len(r.middlewares) - 1; i > -1; i -= 1 {
		m := r.middlewares[i]
		mergeHandler = m(mergeHandler)
	}
	for i := len(middlewares) - 1; i > -1; i -= 1 {
		m := r.middlewares[i]
		mergeHandler = m(mergeHandler)
	}
	r.mux.Handle(path, mergeHandler)
	return r
}

func (r *Router) Serve() error {
	r.srv.Handler = r.mux
	log.Println("Start server")
	return r.srv.ListenAndServe()
}

func (r *Router) Shutdown(ctx context.Context) error {
	log.Println("Shutdown server")
	return r.srv.Shutdown(ctx)
}
