package middleware

import (
	"lb/internal/ratelimiter"
	"log/slog"
	"net"
	"net/http"
)

func WithRateLimiter(rateLimiter *ratelimiter.RateLimiter) Middleware {
	if rateLimiter == nil {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			})
		}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				slog.Error("could not parse remote address", "addr", r.RemoteAddr)
				ip = r.RemoteAddr
			}

			if rateLimitErr := rateLimiter.Hit(ip); rateLimitErr != nil {
				slog.Warn("rate limit exceeded", "err", rateLimitErr)
				w.Header().Set("Retry-After", "1")
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
