package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequireRoles(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userRole       interface{}
		setRole        bool
		allowedRoles   []string
		expectedStatus int
	}{
		{
			name:           "Admin has full access",
			userRole:       "admin",
			setRole:        true,
			allowedRoles:   []string{"viewer"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Viewer has allowed access",
			userRole:       "viewer",
			setRole:        true,
			allowedRoles:   []string{"viewer"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Viewer is denied",
			userRole:       "viewer",
			setRole:        true,
			allowedRoles:   []string{"operator"},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Missing role is denied",
			setRole:        false,
			allowedRoles:   []string{"viewer"},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Invalid role type is denied",
			userRole:       123,
			setRole:        true,
			allowedRoles:   []string{"viewer"},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			router.Use(func(c *gin.Context) {
				if tt.setRole {
					c.Set("user_role", tt.userRole)
				}
				c.Next()
			})

			router.GET("/protected", RequireRoles(tt.allowedRoles...), func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf(
					"expected status %d, got %d",
					tt.expectedStatus,
					rec.Code,
				)
			}
		})
	}
}
