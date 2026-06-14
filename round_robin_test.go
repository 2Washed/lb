package main

import (
	"sync"
	"testing"
)

func TestRoundRobin_NoHealthyServer(t *testing.T) {
	healthyServers.Store([]*Server{})

	server, err := getServer()

	if server != nil {
		t.Error("expected server to be nil")
	}

	if err == nil {
		t.Error("expected error")
	}
}

func TestRoundRobin_SingleServer(t *testing.T) {
	healthyServer := &Server{url: "serverUrl", weight: 1, healthy: true}
	healthyServers.Store([]*Server{healthyServer})

	server, err := getServer()

	if server != healthyServer {
		t.Errorf("expected %s got %s", healthyServer.url, server.url)
	}

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRoundRobin_multiple_servers(t *testing.T) {
	s1 := newTestServer("s1", 1, true)
	s2 := newTestServer("s2", 1, true)

	healthyServers.Store([]*Server{s1, s2})

	expected := []*Server{
		s1, s2, s1, s2, s1, s2,
	}
	callCount := len(expected)

	for i := 0; i < callCount; i++ {
		got, err := getServer()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		want := expected[i]
		if got != want {
			t.Errorf("call %d, expected %s got %s", i+1, want.url, got.url)
		}
	}
}

func TestWeightedRoundRobin_multiple_servers(t *testing.T) {
	s1 := newTestServer("s1", 3, true)
	s2 := newTestServer("s2", 1, true)
	s3 := newTestServer("s3", 2, true)

	healthyServers.Store([]*Server{s1, s2, s3})

	expected := []*Server{
		s1, s1, s1, s2, s3, s3, s1,
	}
	callCount := len(expected)

	for i := 0; i < callCount; i++ {
		got, err := getServer()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		want := expected[i]
		if got != want {
			t.Errorf("call %d, expected %s got %s", i+1, want.url, got.url)
		}
	}
}

func TestWeightedRoundRobin_multiple_servers_concurrent(t *testing.T) {
	s1 := newTestServer("s1", 3, true)
	s2 := newTestServer("s2", 1, true)
	s3 := newTestServer("s3", 2, true)

	healthyServers.Store([]*Server{s1, s2, s3})

	const requestCount = 10_000
	count := map[string]int{}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 0; i < requestCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			s, err := getServer()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			mu.Lock()
			count[s.url]++
			mu.Unlock()
		}()
	}

	wg.Wait()

	if count[s1.url] > int(requestCount*.505) || count[s2.url] > int(requestCount*.167) || count[s3.url] > int(requestCount*.334) {
		t.Errorf("unexpected distribution: %+v", count)
	}
}

func newTestServer(url string, weight int, healthy bool) *Server {
	return &Server{
		url:     url,
		weight:  weight,
		healthy: healthy,
	}
}
