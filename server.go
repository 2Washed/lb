package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type Server struct {
	url                    string
	weight                 int
	healthy                bool
	errorCount             atomic.Int64
	requestCount           atomic.Int64
	activeConnectionsCount atomic.Int64
	mu                     sync.RWMutex
}

func getServer(balancer Balancer) (*Server, error) {
	raw := healthyServers.Load()
	if raw == nil {
		return nil, fmt.Errorf("no healthy servers available")
	}

	servers := raw.([]*Server)
	return balancer.Next(servers)
}

func updateServerHealth(server *Server) {
	isHealthy := isServerHealthy(server)
	server.mu.Lock()
	server.healthy = isHealthy
	server.mu.Unlock()
}
