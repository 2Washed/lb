package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
)

var servers []*Server

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

	go updateHealthyServers(healthCheckDuration)

	http.HandleFunc("/metrics", metricsRequestHandler)
	http.HandleFunc("/", newForwardRequestHandler(maxRetries))
	log.Printf("[INFO] Starting server on port: %v\n", port)
	http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
}

func mapServerConfigToServer(serverConfig *ServerConfiguration) *Server {
	weight := 1
	if serverConfig.Weight > 0 {
		weight = serverConfig.Weight
	}

	return &Server{
		url:    serverConfig.Url,
		weight: weight,
	}
}
