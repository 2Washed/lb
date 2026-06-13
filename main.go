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
	weight  int
	healthy bool
	mu      sync.RWMutex
}

var servers []*Server

var healthyServers atomic.Value
var i atomic.Int64
var rebuildMu sync.Mutex

func main() {
	configuration := getConfiguration()
	port := configuration.Port
	healthCheckDuration := configuration.HealthCheckInterval.Duration
	maxRetries := configuration.MaxRetries
	servers = make([]*Server, 0, len(configuration.Servers))
	for _, serverConfig := range configuration.Servers {
		servers = append(servers, mapServerConfigToServer(&serverConfig))
	}

	for _, server := range servers {
		updateServerHealth(server)
	}
	rebuildHealthyServers()

	go func() {
		for {
			var healthCheckWg sync.WaitGroup
			for _, server := range servers {
				healthCheckWg.Add(1)
				go func(server *Server) {
					defer healthCheckWg.Done()
					updateServerHealth(server)
				}(server)
			}
			healthCheckWg.Wait()
			rebuildHealthyServers()

			time.Sleep(healthCheckDuration)
		}
	}()

	http.HandleFunc("/", newRequestHandler(maxRetries))
	log.Printf("[INFO] Starting server on port: %v\n", port)
	http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
}

func rebuildHealthyServers() {
	// Suppose we have N requests going to the same server and they fail, each request will call this method which will result in N updates to healthyServers
	// Added lock so that only 1 request gets to update healthy servers, while the others will proceed to retry (which can fail if we land on the server actively being marked as unhealthy)
	if !rebuildMu.TryLock() {
		return
	}
	defer rebuildMu.Unlock()

	okServers := filter(servers, func(server *Server) bool {
		server.mu.RLock()
		healthy := server.healthy
		server.mu.RUnlock()
		return healthy
	})
	healthyServers.Store(okServers)

}
func newRequestHandler(maxRetries int) func(http.ResponseWriter, *http.Request) {
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

		res, forwardErr := forwardRequest(req, host.url)
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

			res, forwardErr = forwardRequest(req, host.url)
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

func forwardRequest(req *http.Request, host string) (*http.Response, error) {
	log.Printf("[INFO] Forwarding to %s\n", host)
	outReq := req.Clone(req.Context())
	outReq.URL.Scheme = "http"
	outReq.RequestURI = ""
	outReq.URL.Host = host
	outReq.Host = host

	res, forwardErr := http.DefaultClient.Do(outReq)
	if forwardErr != nil {
		log.Printf("[ERROR] Forwarding request failed, Error: %v\n", forwardErr)
		return nil, fmt.Errorf("forwarding request to host: %s failed", host)
	}

	return res, nil
}

func getServer() (*Server, error) {
	raw := healthyServers.Load()
	if raw == nil {
		return nil, fmt.Errorf("no healthy servers available")
	}

	servers := raw.([]*Server)
	if len(servers) == 0 {
		return nil, fmt.Errorf("no healthy servers available")
	}

	totalWeight := 0
	for _, s := range servers {
		totalWeight += s.weight
	}

	position := int(i.Add(1)-1) % totalWeight

	for _, s := range servers {
		position -= s.weight
		if position < 0 {
			return s, nil
		}
	}

	return nil, fmt.Errorf("no healthy servers available")
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

func mapServerConfigToServer(serverConfig *ServerConfiguration) *Server {
	weight := 1
	if serverConfig.Weight > 1 {
		weight = serverConfig.Weight
	}

	return &Server{
		url:    serverConfig.Url,
		weight: weight,
	}
}
