package metrics

import (
	"encoding/json"
	"lb/internal/server"
	"log"
	"net/http"
	"time"
)

type ServerStats struct {
	Url               string `json:"url"`
	Healthy           bool   `json:"healthy"`
	Requests          int    `json:"requests"`
	Errors            int    `json:"errors"`
	ActiveConnections int    `json:"activeConnections"`
}

type LbStats struct {
	UpTime  string         `json:"uptime"`
	Servers []*ServerStats `json:"servers"`
}

var startTime = time.Now()

func getMetrics(servers []*server.Server) *LbStats {
	metrics := &LbStats{
		UpTime:  time.Since(startTime).Round(time.Second).String(),
		Servers: []*ServerStats{},
	}

	for _, server := range servers {
		metrics.Servers = append(metrics.Servers, getServerStats(server))
	}

	return metrics
}

func getServerStats(s *server.Server) *ServerStats {
	s.Mu.RLock()
	metrics := &ServerStats{
		Url:               s.Url,
		Healthy:           s.Healthy,
		Requests:          int(s.RequestCount.Load()),
		Errors:            int(s.ErrorCount.Load()),
		ActiveConnections: int(s.ActiveConnectionsCount.Load()),
	}
	s.Mu.RUnlock()
	return metrics
}

func NewMetricsRequestHandler(servers []*server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics := getMetrics(servers)

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(metrics)
		if err != nil {
			log.Printf("[ERROR] unexpected error while writing response, error: %v\n", err)
			return
		}
	}
}
