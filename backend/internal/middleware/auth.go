package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/VerryHdyt/cloudforge/backend/internal/models"
	"github.com/VerryHdyt/cloudforge/backend/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		parts := strings.SplitN(authHeader, " ", 2)

		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") ||
			strings.TrimSpace(parts[1]) == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization token is required",
			})
			return
		}

		claims, err := services.ValidateToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			return
		}

		userID, err := strconv.ParseUint(claims.Subject, 10, 64)
		if err != nil || userID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid user identity",
			})
			return
		}

		var user models.User
		if err := db.First(&user, userID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "User account not found or inactive",
				})
				return
			}

			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to verify user account",
			})
			return
		}

		c.Set("user_id", user.ID)
		c.Set("user_role", user.Role)

		c.Next()
	}
}
