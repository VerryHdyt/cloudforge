package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/VerryHdyt/cloudforge/backend/internal/services"
	"github.com/containerd/errdefs"
	"github.com/gin-gonic/gin"
)

func TestHandleDockerError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "container ID required",
			err:            services.ErrContainerIDRequired,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Container ID is required",
		},
		{
			name:           "invalid container ID",
			err:            services.ErrInvalidContainerID,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid container ID",
		},
		{
			name:           "protected container",
			err:            services.ErrProtectedContainer,
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Operation is not allowed for this container",
		},
		{
			name:           "container not found",
			err:            errdefs.ErrNotFound,
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Container not found",
		},
		{
			name:           "operation conflict",
			err:            errdefs.ErrConflict,
			expectedStatus: http.StatusConflict,
			expectedBody:   "Container operation conflicts with its current state",
		},
		{
			name:           "already exists",
			err:            errdefs.ErrAlreadyExists,
			expectedStatus: http.StatusConflict,
			expectedBody:   "Container operation conflicts with its current state",
		},
		{
			name:           "permission denied",
			err:            errdefs.ErrPermissionDenied,
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Operation is not allowed for this container",
		},
		{
			name:           "Docker unavailable",
			err:            errdefs.ErrUnavailable,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Docker Engine operation failed",
		},
		{
			name:           "deadline exceeded",
			err:            context.DeadlineExceeded,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Docker Engine operation failed",
		},
		{
			name:           "unknown error",
			err:            errors.New("internal secret error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "An internal Docker operation error occurred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(recorder)

			handler := &DockerHandler{}
			handler.handleDockerError(context, tt.err)

			if recorder.Code != tt.expectedStatus {
				t.Errorf(
					"expected status %d, got %d",
					tt.expectedStatus,
					recorder.Code,
				)
			}

			if !contains(recorder.Body.String(), tt.expectedBody) {
				t.Errorf(
					"expected response body to contain %q, got %q",
					tt.expectedBody,
					recorder.Body.String(),
				)
			}

			if contains(recorder.Body.String(), "internal secret error") {
				t.Error("internal error details must not be exposed")
			}
		})
	}
}

func contains(value, substring string) bool {
	return len(value) >= len(substring) &&
		(value == substring ||
			len(substring) > 0 &&
				findSubstring(value, substring))
}

func findSubstring(value, substring string) bool {
	for i := 0; i <= len(value)-len(substring); i++ {
		if value[i:i+len(substring)] == substring {
			return true
		}
	}

	return false
}
