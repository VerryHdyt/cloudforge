package services

import (
	"errors"
	"testing"
)

func TestValidateImageReference(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectedErr error
	}{
		{
			name:     "valid image",
			input:    "nginx",
			expected: "nginx",
		},
		{
			name:     "valid image with tag",
			input:    "nginx:latest",
			expected: "nginx:latest",
		},
		{
			name:     "valid image with registry",
			input:    "docker.io/library/nginx:latest",
			expected: "docker.io/library/nginx:latest",
		},
		{
			name:     "valid image digest",
			input:    "nginx@sha256:abcdef123456",
			expected: "nginx@sha256:abcdef123456",
		},
		{
			name:        "empty reference",
			input:       "",
			expectedErr: ErrImageReferenceRequired,
		},
		{
			name:        "whitespace only",
			input:       "   ",
			expectedErr: ErrImageReferenceRequired,
		},
		{
			name:        "leading dash",
			input:       "-nginx",
			expectedErr: ErrInvalidImageReference,
		},
		{
			name:        "contains whitespace",
			input:       "nginx latest",
			expectedErr: ErrInvalidImageReference,
		},
		{
			name:        "contains URL scheme",
			input:       "https://nginx",
			expectedErr: ErrInvalidImageReference,
		},
		{
			name:        "too long",
			input:       string(make([]byte, 256)),
			expectedErr: ErrInvalidImageReference,
		},
		{
			name:     "trims surrounding whitespace",
			input:    " nginx:latest ",
			expected: "nginx:latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateImageReference(tt.input)

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf(
					"expected error %v, got %v",
					tt.expectedErr,
					err,
				)
			}

			if got != tt.expected {
				t.Errorf(
					"expected reference %q, got %q",
					tt.expected,
					got,
				)
			}
		})
	}
}
