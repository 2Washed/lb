package middleware

import (
	"context"
	"crypto/rand"
	"net/http"
)

type contextKey = string

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
