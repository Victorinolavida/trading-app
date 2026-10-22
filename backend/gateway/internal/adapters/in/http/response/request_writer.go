package response

import (
	"net/http"
)

type RequestWriter struct {
	http.ResponseWriter
	req *http.Request
}

func NewRequestWriter(w http.ResponseWriter, r *http.Request) *RequestWriter {
	return &RequestWriter{ResponseWriter: w, req: r}
}

func (rw *RequestWriter) Request() *http.Request {
	return rw.req
}
