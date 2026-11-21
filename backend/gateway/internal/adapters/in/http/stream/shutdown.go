package stream

import (
	"context"
	"net/http"
)

type shutdownKey struct{}

// Middleware makes the server's shutdown signal available to stream handlers.
func Middleware(done <-chan struct{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), shutdownKey{}, done)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Done returns nil outside the gateway server; a nil channel is disabled in a select.
func Done(ctx context.Context) <-chan struct{} {
	done, _ := ctx.Value(shutdownKey{}).(<-chan struct{})
	return done
}
