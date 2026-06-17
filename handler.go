package main

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"time"
)

var httpClient = &http.Client{Timeout: 5 * time.Second} //TODO add to config :)

func forwardRequest(req *http.Request, host *Server) (*http.Response, error) {
	slog.Info("forwarting request", "to", host.url)
	outReq := req.Clone(req.Context())
	outReq.URL.Scheme = "http"
	outReq.RequestURI = ""
	outReq.URL.Host = host.url
	outReq.Host = host.url

	host.activeConnectionsCount.Add(1)
	defer host.activeConnectionsCount.Add(-1)
	host.requestCount.Add(1)

	res, forwardErr := httpClient.Do(outReq)
	if forwardErr != nil {
		host.errorCount.Add(1)
		slog.Error("forwarding request failed", "error", forwardErr)
		return nil, fmt.Errorf("forwarding request to host: %s failed", host.url)
	}

	if res.StatusCode >= 500 && res.StatusCode <= 599 {
		host.errorCount.Add(1)
	}

	return res, nil
}

func newForwardRequestHandler(maxRetries int, balancer Balancer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		host, noServerErr := getServer(balancer)
		if noServerErr != nil {
			slog.Error("no healthy servers")
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		canRetry := false
		switch req.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			canRetry = true
		}

		res, forwardErr := forwardRequest(req, host)
		retries := 1
		for forwardErr != nil && retries < maxRetries && canRetry {
			if res != nil {
				res.Body.Close()
			}

			host.mu.Lock()
			host.healthy = false
			host.mu.Unlock()
			rebuildHealthyServers()

			host, noServerErr = getServer(balancer)
			if noServerErr != nil {
				log.Printf("[ERROR] %v\n", noServerErr)
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}

			res, forwardErr = forwardRequest(req, host)
			retries++
		}

		if forwardErr != nil {
			w.WriteHeader(http.StatusBadGateway)
			return
		}

		defer res.Body.Close()

		for headerName, headerValues := range res.Header {
			for _, v := range headerValues {
				w.Header().Add(headerName, v)
			}
		}
		w.WriteHeader(res.StatusCode)

		_, copyErr := io.Copy(w, res.Body)
		if copyErr != nil {
			slog.Error("copying request failed", "error", copyErr)
			return
		}
	})
}
