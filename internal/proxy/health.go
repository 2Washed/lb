package proxy

import (
	"lb/internal/server"
	"lb/internal/utils"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var HealthyServers atomic.Value //TODO fix this
var rebuildMu sync.Mutex

func UpdateHealthyServers(servers []*server.Server, healthCheckDuration time.Duration) {
	for {
		var healthCheckWg sync.WaitGroup
		for _, s := range servers {
			healthCheckWg.Add(1)
			go func(s *server.Server) {
				defer healthCheckWg.Done()
				UpdateServerHealth(s)
			}(s)
		}
		healthCheckWg.Wait()
		RebuildHealthyServers(servers)

		time.Sleep(healthCheckDuration)
	}
}

func RebuildHealthyServers(servers []*server.Server) {
	// Suppose we have N requests going to the same server and they fail, each request will call this method which will result in N updates to healthyServers
	// Added lock so that only 1 request gets to update healthy servers, while the others will proceed to retry (which can fail if we land on the server actively being marked as unhealthy)
	if !rebuildMu.TryLock() {
		return
	}
	defer rebuildMu.Unlock()

	okServers := utils.Filter(servers, func(server *server.Server) bool {
		server.Mu.RLock()
		healthy := server.Healthy
		server.Mu.RUnlock()
		return healthy
	})
	HealthyServers.Store(okServers)
}

func isServerHealthy(server *server.Server) bool {
	conn, err := net.DialTimeout("tcp", server.Url, 2*time.Second)
	if err != nil {
		return false
	}

	defer conn.Close()
	return true
}

func UpdateServerHealth(server *server.Server) {
	isHealthy := isServerHealthy(server)
	server.Mu.Lock()
	server.Healthy = isHealthy
	server.Mu.Unlock()
}
