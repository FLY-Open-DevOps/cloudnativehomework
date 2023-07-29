package rest

import (
	"context"
	"io"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRouter(t *testing.T) {
	r := New("0.0.0.0:8080")
	r.RegisterMiddleware(idMiddleware)
	r.RegisterMiddleware(logMiddleware)
	var h http.HandlerFunc = func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte("hello"))
	}
	r.RegisterHandler("/", h)
	go func() {
		r.Serve()
	}()
	time.Sleep(10 * time.Millisecond)
	res, err := http.Get("http://localhost:8080")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	log.Println(string(data))
	time.Sleep(2 * time.Second)
	if err := r.Shutdown(context.TODO()); err != nil {
		t.Fatal(err)
	}

}

var (
	logMiddleware Middleware = func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			id := req.Header.Get("id")
			log.Printf("id: %s, Request Method: [%s], from: %s", id, req.Method, req.Host)
			h.ServeHTTP(res, req)
			log.Printf("id: %s, Ending", id)
		})
	}
	idMiddleware Middleware = func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			id := uuid.New().String()
			req.Header.Set("id", id)
			h.ServeHTTP(res, req)
		})
	}
)
