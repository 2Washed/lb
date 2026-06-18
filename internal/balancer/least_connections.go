package balancer

import (
	"fmt"
	"lb/internal/server"
)

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
