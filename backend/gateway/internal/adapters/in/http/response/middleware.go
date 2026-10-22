package response

import "net/http"

func RequestIDResponseMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(NewRequestWriter(w, r), r)
	})
}
