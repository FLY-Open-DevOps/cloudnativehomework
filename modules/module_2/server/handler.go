package server

import "net/http"

type Handler func(*Response, *http.Request)

func (h Handler) Wrap() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h(NewResponse(w), r)
	})
}
