package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

func WithLogging() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			path := ""
			if r.URL != nil {
				path = r.URL.Path
			}
			next.ServeHTTP(w, r)
			slog.Info("new request",
				"method", r.Method,
				"url", path,
				"source", r.RemoteAddr,
				"time", time.Since(start),
				"request-id", getRequestID(r.Context()),
			)
		})
	}
}
