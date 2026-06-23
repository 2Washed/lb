package config_test

import (
	"encoding/json"
	"lb/internal/config"
	"testing"
)

func TestBalancingAlgorithm_unmarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected config.BalancingAlgorithm
		wantErr  bool
	}{
		{
			name:     "round robin",
			input:    `"round-robin"`,
			expected: config.RoundRobin,
			wantErr:  false,
		},
		{
			name:     "least connections",
			input:    `"least-connections"`,
			expected: config.LeastConnections,
			wantErr:  false,
		},
		{
			name:    "invalid",
			input:   `"invalid"`,
			wantErr: true,
		},
		{
			name:    "null",
			input:   "null",
			wantErr: true,
		},
		{
			name:    "number instead of string",
			input:   `123`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b config.BalancingAlgorithm

			err := json.Unmarshal([]byte(tt.input), &b)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if b != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, b)
			}
		})
	}
}

func TestGetBalacingAlgorithm(t *testing.T) {
	tests := []struct {
		name     string
		expected config.BalancingAlgorithm
		ok       bool
	}{
		{"round-robin", config.RoundRobin, true},
		{"least-connections", config.LeastConnections, true},
		{"invalid", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := config.GetBalacingAlgorithm(tt.name)

			if ok != tt.ok {
				t.Fatalf("expected ok=%v, got %v", tt.ok, ok)
			}

			if ok && got != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestGetBalancingAlgorithmName(t *testing.T) {
	tests := []struct {
		algo     config.BalancingAlgorithm
		expected string
	}{
		{config.RoundRobin, "round-robin"},
		{config.LeastConnections, "least-connections"},
	}

	for _, tt := range tests {
		got := config.GetBalancingAlgorithmName(tt.algo)

		if got != tt.expected {
			t.Fatalf("expected %q, got %q", tt.expected, got)
		}
	}
}
