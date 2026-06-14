package main

import (
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var healthyServers atomic.Value

func updateHealthyServers(healthCheckDuration time.Duration) {
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

func isServerHealthy(server *Server) bool {
	conn, err := net.DialTimeout("tcp", server.url, 2*time.Second)
	if err != nil {
		return false
	}

	defer conn.Close()
	return true
}
