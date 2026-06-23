package main

import (
	"fmt"
	"lb/internal/balancer"
	"lb/internal/config"
	"lb/internal/metrics"
	"lb/internal/middleware"
	"lb/internal/proxy"
	"lb/internal/ratelimiter"
	"lb/internal/server"
	"log/slog"
	"net/http"
	"time"
)

var servers []*server.Server

func main() {
	configuration := config.GetConfiguration()

	port := configuration.Port
	healthCheckDuration := configuration.HealthCheckInterval.Duration
	maxRetries := configuration.MaxRetries
	servers = make([]*server.Server, 0, len(configuration.Servers))
	for _, serverConfig := range configuration.Servers {
		servers = append(servers, mapServerConfigToServer(serverConfig))
	}
	balancer := balancingAlgoToBalancer(configuration.BalancingAlgorithm)

	var rl *ratelimiter.RateLimiter
	if configuration.RateLimiter != nil {
		rl = ratelimiter.NewRateLimiter(
			configuration.RateLimiter.Rate,
			configuration.RateLimiter.BurstSeconds,
			configuration.RateLimiter.Expiry.Duration,
		)
	}

	for _, server := range servers {
		proxy.UpdateServerHealth(server)
	}
	proxy.RebuildHealthyServers(servers)

	go proxy.UpdateHealthyServers(servers, healthCheckDuration)

	http.HandleFunc("/metrics", metrics.NewMetricsRequestHandler(servers))

	var httpClient = &http.Client{Timeout: 5 * time.Second} //TODO add to config :)

	handler := middleware.Chain(
		proxy.NewForwardRequestHandler(maxRetries, balancer, servers, httpClient),
		middleware.WithRateLimiter(rl),
		middleware.WithLogging(),
		middleware.WithRequestID(),
		middleware.WithRecover(),
	)
	http.Handle("/", handler)

	slog.Info(
		"starting server",
		"port", port,
		"healthCheckDuration", healthCheckDuration,
		"maxRetries", maxRetries,
		"balancer", config.GetBalancingAlgorithmName(configuration.BalancingAlgorithm),
	)

	err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil)

	if err != nil {
		slog.Error("error while starting server", "err", err)
	}
}

func mapServerConfigToServer(serverConfig *config.ServerConfiguration) *server.Server {
	weight := 1
	if serverConfig.Weight > 0 {
		weight = serverConfig.Weight
	}

	return &server.Server{
		Url:    serverConfig.Url,
		Weight: weight,
	}
}

func balancingAlgoToBalancer(algo config.BalancingAlgorithm) balancer.Balancer {
	switch algo {
	case config.RoundRobin:
		return &balancer.RoundRobin{}
	case config.LeastConnections:
		return &balancer.LeastConnections{}
	}

	panic("unreachable")
}
