package middleware

import (
	"net/http"
)

type Middleware = func(http.Handler) http.Handler

func Chain(baseHandler http.Handler, middlewares ...Middleware) http.Handler {
	handler := baseHandler
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}

	return handler
}
