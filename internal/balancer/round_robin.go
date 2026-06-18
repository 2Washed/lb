package balancer

import (
	"fmt"
	"lb/internal/server"
	"sync/atomic"
)

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
