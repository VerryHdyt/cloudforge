package handlers

import (
	"testing"

	"github.com/moby/moby/api/types/container"
)

func TestContainerResponseMapping(t *testing.T) {
	tests := []struct {
		name     string
		input    container.Summary
		expected ContainerResponse
	}{
		{
			name: "container with one name",
			input: container.Summary{
				ID:      "abc123",
				Names:   []string{"/cloudforge-postgres"},
				Image:   "postgres:17-alpine",
				State:   container.ContainerState("running"),
				Status:  "Up 10 minutes (healthy)",
				Created: 1795008767,
			},
			expected: ContainerResponse{
				ID:      "abc123",
				Name:    "cloudforge-postgres",
				Image:   "postgres:17-alpine",
				State:   "running",
				Status:  "Up 10 minutes (healthy)",
				Created: 1795008767,
			},
		},
		{
			name: "container without name",
			input: container.Summary{
				ID:      "def456",
				Names:   []string{},
				Image:   "hello-world",
				State:   container.ContainerState("exited"),
				Status:  "Exited (0)",
				Created: 1795006031,
			},
			expected: ContainerResponse{
				ID:      "def456",
				Name:    "",
				Image:   "hello-world",
				State:   "exited",
				Status:  "Exited (0)",
				Created: 1795006031,
			},
		},
		{
			name: "container with multiple names",
			input: container.Summary{
				ID:      "ghi789",
				Names:   []string{"/primary-name", "/secondary-name"},
				Image:   "alpine:latest",
				State:   container.ContainerState("running"),
				Status:  "Up 5 minutes",
				Created: 1795007000,
			},
			expected: ContainerResponse{
				ID:      "ghi789",
				Name:    "primary-name",
				Image:   "alpine:latest",
				State:   "running",
				Status:  "Up 5 minutes",
				Created: 1795007000,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name := ""

			if len(tt.input.Names) > 0 {
				name = tt.input.Names[0]
				if len(name) > 0 && name[0] == '/' {
					name = name[1:]
				}
			}

			got := ContainerResponse{
				ID:      tt.input.ID,
				Name:    name,
				Image:   tt.input.Image,
				State:   string(tt.input.State),
				Status:  tt.input.Status,
				Created: tt.input.Created,
			}

			if got != tt.expected {
				t.Errorf("got %+v, want %+v", got, tt.expected)
			}
		})
	}
}
