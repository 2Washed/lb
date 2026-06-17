package main

import (
	"fmt"
	"log/slog"
	"net/http"
)

var servers []*Server

func main() {
	configuration := getConfiguration()

	port := configuration.Port
	healthCheckDuration := configuration.HealthCheckInterval.Duration
	maxRetries := configuration.MaxRetries
	servers = make([]*Server, 0, len(configuration.Servers))
	for _, serverConfig := range configuration.Servers {
		servers = append(servers, mapServerConfigToServer(serverConfig))
	}
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

	handler := Chain(
		newForwardRequestHandler(maxRetries, balancer),
		WithRateLimiter(rateLimiter),
		WithLogging(),
		WithRequestID(),
		WithRecover(),
	)
	http.Handle("/", handler)
	slog.Info("starting server", "port", port, "healthCheckDuration", healthCheckDuration, "maxRetries", maxRetries, "balancer", algoToString[configuration.BalancingAlgorithm])
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
