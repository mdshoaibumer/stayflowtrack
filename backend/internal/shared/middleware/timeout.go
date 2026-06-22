package middleware

import (
	"context"
	"net/http"
	"time"
)

// RequestTimeout wraps each request with a context deadline to prevent runaway operations.
// The timeout is applied at the handler level, so all downstream DB queries, external calls,
// etc. will be cancelled when the deadline expires.
func RequestTimeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
