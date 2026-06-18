package middleware

import (
	"context"
	"crypto/rand"
	"lb/internal/ratelimiter"
	"log/slog"
	"net"
	"net/http"
	"time"
)

type Middleware = func(http.Handler) http.Handler
type contextKey = string

func Chain(baseHandler http.Handler, middlewares ...Middleware) http.Handler {
	handler := baseHandler
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}

	return handler
}

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

const REQUEST_ID_KEY contextKey = "requestId"

func WithRequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestId := rand.Text()
			ctx := context.WithValue(r.Context(), REQUEST_ID_KEY, requestId)
			w.Header().Set("X-Request-ID", requestId)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func getRequestID(ctx context.Context) string {
	id, _ := ctx.Value(REQUEST_ID_KEY).(string)
	return id
}

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
