package test

import (
	"lb/internal/balancer"
	"lb/internal/server"
	"testing"
)

func TestLC_NoHealthyServer(t *testing.T) {
	rb := balancer.LeastConnections{}
	servers := []*server.Server{}

	server, err := rb.Next(servers)

	if server != nil {
		t.Error("expected server to be nil")
	}

	if err == nil {
		t.Error("expected error")
	}
}

func TestLC_SingleServer(t *testing.T) {
	rb := balancer.LeastConnections{}
	healthyServer := newTestServer("s1", 1, true)
	servers := []*server.Server{healthyServer}

	server, err := rb.Next(servers)

	if server != healthyServer {
		t.Errorf("expected %s got %s", healthyServer.Url, server.Url)
	}

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLC_multiple_servers(t *testing.T) {
	rb := balancer.LeastConnections{}

	s1 := newTestServer("s1", 1, true)
	s1.ActiveConnectionsCount.Store(1)
	s2 := newTestServer("s2", 1, true)
	s2.ActiveConnectionsCount.Store(2)

	servers := []*server.Server{s1, s2}

	expected := []*server.Server{
		s1, s1, s2, s1, s2, s1,
	}
	callCount := len(expected)

	for i := 0; i < callCount; i++ {
		got, err := rb.Next(servers)
		got.ActiveConnectionsCount.Add(1)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		want := expected[i]
		if got != want {
			t.Errorf("call %d, expected %s got %s", i+1, want.Url, got.Url)
		}
	}
}
