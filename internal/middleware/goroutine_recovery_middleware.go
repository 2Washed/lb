package middleware

import (
	"log/slog"
	"net/http"
)

func WithRecover() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					slog.Error("panic recovered", "err", err, "path", r.URL.Path)
					w.WriteHeader(http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
