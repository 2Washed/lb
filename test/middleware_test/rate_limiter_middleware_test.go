package middleware_test

import (
	"lb/internal/middleware"
	"lb/internal/ratelimiter"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterMiddleware_should_forward_request_if_no_rl_defined(t *testing.T) {
	nextCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.WithRateLimiter(nil)(testHandler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !nextCalled {
		t.Errorf("expected next handler to be called")
	}

	if rr.Code != http.StatusOK {
		t.Errorf("expected %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestRateLimiterMiddleware_should_forward_request_if_rate_limit_ok(t *testing.T) {
	nextCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})
	rateLimiter := ratelimiter.NewRateLimiter(1, 1, 10*time.Second)

	handler := middleware.WithRateLimiter(rateLimiter)(testHandler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !nextCalled {
		t.Errorf("expected next handler to be called")
	}

	if rr.Code != http.StatusOK {
		t.Errorf("expected %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestRateLimiter_should_not_forward_request_if_rate_limit_ko(t *testing.T) {
	nextCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})
	rateLimiter := ratelimiter.NewRateLimiter(0, 0, 10*time.Second)

	handler := middleware.WithRateLimiter(rateLimiter)(testHandler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if nextCalled {
		t.Errorf("expected next handler to not be called")
	}

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("expected %d, got %d", http.StatusTooManyRequests, rr.Code)
	}
}
