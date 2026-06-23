package config_test

import (
	"encoding/json"
	"lb/internal/config"
	"testing"
	"time"
)

func TestDuration_unmarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{
			name:     "seconds",
			input:    `"10s"`,
			expected: 10 * time.Second,
		},
		{
			name:     "minutes and seconds",
			input:    `"1m30s"`,
			expected: 90 * time.Second,
		},
		{
			name:     "hours",
			input:    `"2h"`,
			expected: 2 * time.Hour,
		},
		{
			name:    "invalid duration",
			input:   `"abc"`,
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   `""`,
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
			var d config.Duration

			err := json.Unmarshal([]byte(tt.input), &d)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if d.Duration != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, d.Duration)
			}
		})
	}
}
