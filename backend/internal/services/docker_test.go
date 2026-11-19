package services

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	containerTypes "github.com/moby/moby/api/types/container"
)

func TestValidateContainerID(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectedErr error
	}{
		{
			name:     "valid container name",
			input:    "cloudforge-api",
			expected: "cloudforge-api",
		},
		{
			name:     "valid hexadecimal ID",
			input:    "abc123def456",
			expected: "abc123def456",
		},
		{
			name:     "valid name with underscore",
			input:    "cloudforge_api",
			expected: "cloudforge_api",
		},
		{
			name:        "empty ID",
			input:       "",
			expectedErr: ErrContainerIDRequired,
		},
		{
			name:        "whitespace only",
			input:       "   ",
			expectedErr: ErrContainerIDRequired,
		},
		{
			name:        "ID containing spaces",
			input:       "cloudforge api",
			expectedErr: ErrInvalidContainerID,
		},
		{
			name:        "ID containing slash",
			input:       "cloudforge/api",
			expectedErr: ErrInvalidContainerID,
		},
		{
			name:        "ID containing special characters",
			input:       "cloudforge@api",
			expectedErr: ErrInvalidContainerID,
		},
		{
			name:        "ID longer than 128 characters",
			input:       strings.Repeat("a", 129),
			expectedErr: ErrInvalidContainerID,
		},
		{
			name:     "ID with surrounding whitespace",
			input:    " cloudforge-api ",
			expected: "cloudforge-api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateContainerID(tt.input)

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}

			if got != tt.expected {
				t.Errorf("expected ID %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestCalculateContainerStats(t *testing.T) {
	input := `{
		"id": "abc123",
		"name": "/cloudforge-api",
		"read": "2026-11-19T13:42:31Z",
		"cpu_stats": {
			"cpu_usage": {
				"total_usage": 300,
				"percpu_usage": [100, 200]
			},
			"system_cpu_usage": 1000,
			"online_cpus": 2
		},
		"precpu_stats": {
			"cpu_usage": {
				"total_usage": 100
			},
			"system_cpu_usage": 500
		},
		"memory_stats": {
			"usage": 512,
			"limit": 1024
		},
		"networks": {
			"eth0": {
				"rx_bytes": 100,
				"tx_bytes": 200
			},
			"eth1": {
				"rx_bytes": 300,
				"tx_bytes": 400
			}
		},
		"blkio_stats": {
			"io_service_bytes_recursive": [
				{
					"op": "Read",
					"value": 1000
				},
				{
					"op": "Write",
					"value": 2000
				},
				{
					"op": "Sync",
					"value": 500
				}
			]
		}
	}`

	var stats containerTypes.StatsResponse

	if err := json.Unmarshal([]byte(input), &stats); err != nil {
		t.Fatalf("failed to decode test stats: %v", err)
	}

	got := calculateContainerStats(stats)

	if got.ContainerID != "abc123" {
		t.Errorf("expected container ID abc123, got %q", got.ContainerID)
	}

	if got.Name != "cloudforge-api" {
		t.Errorf("expected name cloudforge-api, got %q", got.Name)
	}

	if got.CPUPercent != 80 {
		t.Errorf("expected CPU percent 80, got %f", got.CPUPercent)
	}

	if got.MemoryUsage != 512 {
		t.Errorf("expected memory usage 512, got %d", got.MemoryUsage)
	}

	if got.MemoryLimit != 1024 {
		t.Errorf("expected memory limit 1024, got %d", got.MemoryLimit)
	}

	if got.MemoryPercent != 50 {
		t.Errorf("expected memory percent 50, got %f", got.MemoryPercent)
	}

	if got.NetworkRx != 400 {
		t.Errorf("expected network RX 400, got %d", got.NetworkRx)
	}

	if got.NetworkTx != 600 {
		t.Errorf("expected network TX 600, got %d", got.NetworkTx)
	}

	if got.BlockRead != 1000 {
		t.Errorf("expected block read 1000, got %d", got.BlockRead)
	}

	if got.BlockWrite != 2000 {
		t.Errorf("expected block write 2000, got %d", got.BlockWrite)
	}
}

func TestCalculateContainerStatsZeroValues(t *testing.T) {
	stats := containerTypes.StatsResponse{}

	got := calculateContainerStats(stats)

	if got.CPUPercent != 0 {
		t.Errorf("expected CPU percent 0, got %f", got.CPUPercent)
	}

	if got.MemoryPercent != 0 {
		t.Errorf("expected memory percent 0, got %f", got.MemoryPercent)
	}

	if got.MemoryUsage != 0 {
		t.Errorf("expected memory usage 0, got %d", got.MemoryUsage)
	}

	if got.NetworkRx != 0 || got.NetworkTx != 0 {
		t.Errorf(
			"expected network RX/TX 0, got RX=%d TX=%d",
			got.NetworkRx,
			got.NetworkTx,
		)
	}

	if got.BlockRead != 0 || got.BlockWrite != 0 {
		t.Errorf(
			"expected block read/write 0, got read=%d write=%d",
			got.BlockRead,
			got.BlockWrite,
		)
	}
}
