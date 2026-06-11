package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type Server struct {
	url     string
	healthy bool
	mu      sync.RWMutex
}

var servers = []*Server{
	{
		url: "127.0.0.1:9000",
	},
	{
		url: "127.0.0.1:9001",
	},
	{
		url: "127.0.0.1:9002",
	},
}

var healthyServers atomic.Value

const HEALTH_CHECK_DURATION_SECONDS = 10

var i atomic.Int64

func main() {
	healthyServers.Store(servers)

	go func() {
		for {
			for _, server := range servers {
				updateServerHealth(server)
			}
			okServers := filter(servers, func(server *Server) bool {
				server.mu.RLock()
				healthy := server.healthy
				server.mu.RUnlock()
				return healthy
			})
			healthyServers.Store(okServers)

			time.Sleep(HEALTH_CHECK_DURATION_SECONDS)
		}
	}()

	http.HandleFunc("/", requestHandler)
	http.ListenAndServe(":8080", nil)
}

func requestHandler(w http.ResponseWriter, req *http.Request) {
	//Clone request
	outReq := req.Clone(req.Context())
	outReq.URL.Scheme = "http"
	outReq.RequestURI = ""

	host, noServerErr := getServer()
	if noServerErr != nil {
		log.Printf("[ERROR] %v\n", noServerErr)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	outReq.URL.Host = host
	outReq.Host = host

	//Forward the request
	res, forwardErr := http.DefaultClient.Do(outReq)
	if forwardErr != nil {
		log.Printf("[ERROR] Forwarding request failed, Error: %v\n", forwardErr)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(nil)
		return
	}
	defer res.Body.Close()

	//Write headers
	for headerName, headerValues := range res.Header {
		for _, v := range headerValues {
			w.Header().Add(headerName, v)
		}
	}
	w.WriteHeader(res.StatusCode)

	_, copyErr := io.Copy(w, res.Body)
	if copyErr != nil {
		log.Printf("[ERROR] Copying request to client failed, Error: %v\n", copyErr)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(nil)
		return
	}
}

func getServer() (string, error) {
	raw := healthyServers.Load()
	if raw == nil {
		return "", fmt.Errorf("no healthy servers available")
	}

	servers := raw.([]*Server)
	if len(servers) == 0 {
		return "", fmt.Errorf("no healthy servers available")
	}

	idx := i.Add(1)
	serverIndex := int(idx) % len(servers)
	serverUrl := servers[serverIndex].url
	return serverUrl, nil
}

func updateServerHealth(server *Server) {
	isHealthy := isServerHealthy(server)
	server.mu.Lock()
	server.healthy = isHealthy
	server.mu.Unlock()
}

func isServerHealthy(server *Server) bool {
	conn, err := net.DialTimeout("tcp", server.url, 2*time.Second)
	if err != nil {
		return false
	}

	defer conn.Close()
	return true
}
