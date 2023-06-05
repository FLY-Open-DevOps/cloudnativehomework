package server

import "net/http"

type Response struct {
	w    http.ResponseWriter
	code int
}

func NewResponse(resp http.ResponseWriter) *Response {
	return &Response{
		w:    resp,
		code: http.StatusOK,
	}
}

func (r *Response) SetStatusCode(code int) {
	r.code = code
}

func (r *Response) SetHeader(key, value string) {
	r.w.Header().Set(key, value)
}

func (r *Response) StatusCode() int {
	return r.code
}

func (r *Response) Write(data []byte) {
	r.w.Write(data)
}
