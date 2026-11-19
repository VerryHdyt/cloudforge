package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireRoles(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleValue, exists := c.Get("user_role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "User role not found",
			})
			return
		}

		userRole, ok := roleValue.(string)
		if !ok || userRole == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Invalid user role",
			})
			return
		}

		// Admin has full access
		if userRole == "admin" {
			c.Next()
			return
		}

		// Check allowed roles
		for _, role := range allowedRoles {
			if userRole == role {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "You do not have permission to access this resource",
		})
	}
}
