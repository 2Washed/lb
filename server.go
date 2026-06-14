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
