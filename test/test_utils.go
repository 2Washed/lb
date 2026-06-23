package test

import "lb/internal/server"

func NewTestServer(url string, weight int, healthy bool) *server.Server {
	return &server.Server{
		Url:     url,
		Weight:  weight,
		Healthy: healthy,
	}
}
