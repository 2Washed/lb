package proxy_test

import (
	"fmt"
	"lb/internal/proxy"
	"lb/internal/server"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testBalancer struct {
	err        bool
	mockServer *server.Server
}

func (tb *testBalancer) Next(servers []*server.Server) (*server.Server, error) {
	if tb.err {
		return nil, fmt.Errorf("no healthy servers")
	} else {
		return tb.mockServer, nil
	}
}

type RoundTripFunc func(*http.Request) (*http.Response, error)

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestForwardRequestHandler_should_return_service_unavailable_when_no_healthy_servers(t *testing.T) {
	servers := []*server.Server{
		{
			Url: "s1",
		},
	}
	balancer := &testBalancer{
		err: true,
	}
	mockClient := &http.Client{}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler := proxy.NewForwardRequestHandler(1, balancer, servers, mockClient)

	handler.ServeHTTP(rr, req)

	expectedCode := http.StatusServiceUnavailable
	gotCode := rr.Code
	if gotCode != expectedCode {
		t.Fatalf("unexpected status code, expected %v got %v", expectedCode, gotCode)
	}
}

func TestForwardRequestHandler_should_retry_on_idempotent_methods(t *testing.T) {
	s := &server.Server{
		Url: "127.0.0.1:80",
	}
	balancer := &testBalancer{
		err:        false,
		mockServer: s,
	}

	tests := []struct {
		name        string
		method      string
		expectRetry bool
	}{
		{
			name:        "GET",
			method:      http.MethodGet,
			expectRetry: true,
		},
		{
			name:        "OPTIONS",
			method:      http.MethodOptions,
			expectRetry: true,
		},
		{
			name:        "HEAD",
			method:      http.MethodHead,
			expectRetry: true,
		},
		{
			name:        "POST",
			method:      http.MethodPost,
			expectRetry: false,
		},
		{
			name:        "DELETE",
			method:      http.MethodDelete,
			expectRetry: false,
		},
	}

	for _, tt := range tests {
		proxy.HealthyServers.Store([]*server.Server{s})
		callCount := 0
		maxTries := 2

		req := httptest.NewRequest(tt.method, "/fail", nil)
		mockClient := &http.Client{
			Transport: RoundTripFunc(func(req *http.Request) (*http.Response, error) {
				callCount++
				return nil, fmt.Errorf("fail")
			}),
		}
		handler := proxy.NewForwardRequestHandler(maxTries, balancer, []*server.Server{s}, mockClient)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if tt.expectRetry {
			if callCount != maxTries {
				t.Fatalf("expected %v tries, got %v", maxTries, callCount)
			}
		} else {
			if callCount > 1 {
				t.Fatalf("expected 1 try, got %v", callCount)
			}
		}

	}
}
