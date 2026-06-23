package middleware_test

import (
	"lb/internal/middleware"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIdMiddleware(t *testing.T) {
	var requestIdFromContext string
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestIdFromContext = middleware.GetRequestID(r.Context())

		if requestIdFromContext == "" {
			t.Fatal("expected request ID in context")
		}

		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.WithRequestID()(testHandler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	header := rr.Header().Get("X-Request-ID")
	if header == "" {
		t.Fatal("expected header ('X-Request-ID') to be set")
	}

	if header != requestIdFromContext {
		t.Fatalf("header does not match id, expected %s got %s", requestIdFromContext, header)
	}
}
