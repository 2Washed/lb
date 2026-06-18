package balancer

import (
	"lb/internal/server"
)

type Balancer interface {
	Next(servers []*server.Server) (*server.Server, error)
}
