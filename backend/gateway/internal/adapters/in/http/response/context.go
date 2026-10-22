package response

import (
	"context"
	"net/http"
)

type ctxKey string

const requestKey ctxKey = "http_request"

func WithRequest(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, requestKey, r)
}

func RequestFromContext(w http.ResponseWriter) *http.Request {
	if r, ok := w.(requester); ok {
		return r.Request()
	}
	if rc, ok := w.(interface {
		Unwrap() http.ResponseWriter
	}); ok {
		// not used
		_ = rc
	}
	return nil
}

type requester interface {
	Request() *http.Request
}
