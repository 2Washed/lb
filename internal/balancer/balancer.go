package balancer

import (
	"fmt"
	"lb/internal/server"
	"sync/atomic"
)

type Balancer interface {
	Next(servers []*server.Server) (*server.Server, error)
}

type RoundRobin struct {
	counter atomic.Int64
}

func (r *RoundRobin) Next(servers []*server.Server) (*server.Server, error) {
	if len(servers) == 0 {
		return nil, fmt.Errorf("no healthy servers available")
	}

	totalWeight := 0
	for _, s := range servers {
		totalWeight += s.Weight
	}

	position := int(r.counter.Add(1)-1) % totalWeight

	for _, s := range servers {
		position -= s.Weight
		if position < 0 {
			return s, nil
		}
	}

	panic("unreachable")
}

type LeastConnections struct{}

func (l *LeastConnections) Next(servers []*server.Server) (*server.Server, error) {
	if len(servers) == 0 {
		return nil, fmt.Errorf("no healthy servers available")
	}

	//This will cause the first server connections to spike on a cold start -Claude
	best := servers[0]
	for _, s := range servers[1:] {
		if s.ActiveConnectionsCount.Load() < best.ActiveConnectionsCount.Load() {
			best = s
		}
	}
	return best, nil
}
