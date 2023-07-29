package rest

// type Handler func(*Response, Request)

// type HandlerStruct struct {
// 	Handler Handler
// 	Path    string
// }

// // correct here to solve stauts code lost
// func (h Handler) Wrap() http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		h(NewResponse(w), Request(r))
// 	})
// }

// func (h Handler) Serve(res *Response, req Request) {
// 	h.Wrap().ServeHTTP(res.w, req)
// }

// type Response struct {
// 	w    http.ResponseWriter
// 	code int
// }

// type Request *http.Request

// func NewResponse(resp http.ResponseWriter) *Response {
// 	return &Response{
// 		w:    resp,
// 		code: http.StatusOK,
// 	}
// }

// func (r *Response) SetStatusCode(code int) {
// 	r.w.WriteHeader(code)
// 	r.code = code
// }

// func (r *Response) SetHeader(key, value string) {
// 	r.w.Header().Set(key, value)
// }

// func (r *Response) StatusCode() int {
// 	return r.code
// }

// func (r *Response) Write(data []byte) {
// 	r.w.Write(data)
// }

// func (r *Response) SetContentType(value string) {
// 	r.w.Header().Set("Content-Type", value)
// }
