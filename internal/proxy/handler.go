package proxy

import (
	"fmt"
	"io"
	"lb/internal/balancer"
	"lb/internal/server"
	"log"
	"log/slog"
	"net/http"
	"time"
)

var httpClient = &http.Client{Timeout: 5 * time.Second} //TODO add to config :)

func forwardRequest(req *http.Request, host *server.Server) (*http.Response, error) {
	slog.Info("forwarding request", "to", host.Url)
	outReq := req.Clone(req.Context())
	outReq.URL.Scheme = "http"
	outReq.RequestURI = ""
	outReq.URL.Host = host.Url
	outReq.Host = host.Url

	host.ActiveConnectionsCount.Add(1)
	defer host.ActiveConnectionsCount.Add(-1)
	host.RequestCount.Add(1)

	res, forwardErr := httpClient.Do(outReq)
	if forwardErr != nil {
		host.ErrorCount.Add(1)
		slog.Error("forwarding request failed", "error", forwardErr)
		return nil, fmt.Errorf("forwarding request to host: %s failed", host.Url)
	}

	if res.StatusCode >= 500 && res.StatusCode <= 599 {
		host.ErrorCount.Add(1)
	}

	return res, nil
}

func NewForwardRequestHandler(maxRetries int, balancer balancer.Balancer, servers []*server.Server) http.Handler {
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

			host.Mu.Lock()
			host.Healthy = false
			host.Mu.Unlock()
			RebuildHealthyServers(servers)

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

func getServer(balancer balancer.Balancer) (*server.Server, error) {
	raw := healthyServers.Load()
	if raw == nil {
		return nil, fmt.Errorf("no healthy servers available")
	}

	servers := raw.([]*server.Server)
	return balancer.Next(servers)
}
