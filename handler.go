package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func forwardRequest(req *http.Request, host *Server) (*http.Response, error) {
	log.Printf("[INFO] Forwarding to %s\n", host.url)
	outReq := req.Clone(req.Context())
	outReq.URL.Scheme = "http"
	outReq.RequestURI = ""
	outReq.URL.Host = host.url
	outReq.Host = host.url

	host.activeConnectionsCount.Add(1)
	defer host.activeConnectionsCount.Add(-1)
	host.requestCount.Add(1)

	res, forwardErr := http.DefaultClient.Do(outReq)
	if forwardErr != nil {
		host.errorCount.Add(1)
		log.Printf("[ERROR] Forwarding request failed, Error: %v\n", forwardErr)
		return nil, fmt.Errorf("forwarding request to host: %s failed", host)
	}

	if res.StatusCode >= 500 && res.StatusCode <= 599 {
		host.errorCount.Add(1)
	}

	return res, nil
}

func newForwardRequestHandler(maxRetries int) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		log.Printf("[INFO] New %v request from %v \n", req.Method, req.RemoteAddr)
		host, noServerErr := getServer()
		if noServerErr != nil {
			log.Printf("[ERROR] %v\n", noServerErr)
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
			host.mu.Lock()
			host.healthy = false
			host.mu.Unlock()
			rebuildHealthyServers()

			log.Println("RETRYING")

			host, noServerErr = getServer()
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
			log.Printf("[ERROR] Copying request to client failed, Error: %v\n", copyErr)
			return
		}
	}
}
