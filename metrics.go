package main

import (
	"encoding/json"
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

func getMetrics() *LbStats {
	metrics := &LbStats{
		UpTime:  time.Since(startTime).Round(time.Second).String(),
		Servers: []*ServerStats{},
	}

	for _, server := range servers {
		metrics.Servers = append(metrics.Servers, getServerStats(server))
	}

	return metrics
}

func getServerStats(server *Server) *ServerStats {
	server.mu.RLock()
	metrics := &ServerStats{
		Url:               server.url,
		Healthy:           server.healthy,
		Requests:          int(server.requestCount.Load()),
		Errors:            int(server.errorCount.Load()),
		ActiveConnections: int(server.activeConnectionsCount.Load()),
	}
	server.mu.RUnlock()
	return metrics
}

func metricsRequestHandler(w http.ResponseWriter, _ *http.Request) {
	metrics := getMetrics()

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(metrics)
	if err != nil {
		log.Printf("[ERROR] unexpected error while writing response, error: %v\n", err)
		return
	}
}
