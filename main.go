package main

import (
	"fmt"
	"log"
	"net/http"
)

var servers []*Server

func main() {
	configuration := getConfiguration()

	port := configuration.Port
	log.Printf("using port: %d\n", port)
	healthCheckDuration := configuration.HealthCheckInterval.Duration
	maxRetries := configuration.MaxRetries
	servers = make([]*Server, 0, len(configuration.Servers))
	for _, serverConfig := range configuration.Servers {
		log.Printf("adding server: %v\n", serverConfig)
		servers = append(servers, mapServerConfigToServer(serverConfig))
	}

	log.Printf("using algorithm: %s\n", algoToString[configuration.BalancingAlgorithm])
	balancer := balancingAlgoToBalancer(configuration.BalancingAlgorithm)

	var rateLimiter *RateLimiter
	if configuration.RateLimiter != nil {
		rateLimiter = NewRateLimiter(
			configuration.RateLimiter.Rate,
			configuration.RateLimiter.BurstSeconds,
			configuration.RateLimiter.Expiry.Duration,
		)
	}

	for _, server := range servers {
		updateServerHealth(server)
	}
	rebuildHealthyServers()

	go updateHealthyServers(healthCheckDuration)

	http.HandleFunc("/metrics", metricsRequestHandler)
	http.HandleFunc("/", newForwardRequestHandler(maxRetries, balancer, rateLimiter))
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

func balancingAlgoToBalancer(algo BalancingAlgorithm) Balancer {
	switch algo {
	case RoundRobinAlgo:
		return &RoundRobin{}
	case LeastConnectionsAlgo:
		return &LeastConnections{}
	}

	panic("unreachable")
}
