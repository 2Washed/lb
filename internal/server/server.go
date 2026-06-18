package server

import (
	"sync"
	"sync/atomic"
)

type Server struct {
	Url                    string
	Weight                 int
	Healthy                bool
	ErrorCount             atomic.Int64
	RequestCount           atomic.Int64
	ActiveConnectionsCount atomic.Int64
	Mu                     sync.RWMutex
}
