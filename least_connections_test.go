package main

import (
	"testing"
)

func TestLC_NoHealthyServer(t *testing.T) {
	rb := LeastConnections{}
	servers := []*Server{}

	server, err := rb.Next(servers)

	if server != nil {
		t.Error("expected server to be nil")
	}

	if err == nil {
		t.Error("expected error")
	}
}

func TestLC_SingleServer(t *testing.T) {
	rb := LeastConnections{}
	healthyServer := newTestServer("s1", 1, true)
	servers := []*Server{healthyServer}

	server, err := rb.Next(servers)

	if server != healthyServer {
		t.Errorf("expected %s got %s", healthyServer.url, server.url)
	}

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLC_multiple_servers(t *testing.T) {
	rb := LeastConnections{}

	s1 := newTestServer("s1", 1, true)
	s1.activeConnectionsCount.Store(1)
	s2 := newTestServer("s2", 1, true)
	s2.activeConnectionsCount.Store(2)

	servers := []*Server{s1, s2}

	expected := []*Server{
		s1, s1, s2, s1, s2, s1,
	}
	callCount := len(expected)

	for i := 0; i < callCount; i++ {
		got, err := rb.Next(servers)
		got.activeConnectionsCount.Add(1)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		want := expected[i]
		if got != want {
			t.Errorf("call %d, expected %s got %s", i+1, want.url, got.url)
		}
	}
}
