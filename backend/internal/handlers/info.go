package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func PlatformInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"name":        "CloudForge",
		"description": "Open-source cloud infrastructure management platform",
		"version":     "0.1.0",
		"status":      "development",
		"stack": []string{
			"Go",
			"React",
			"PostgreSQL",
			"Docker",
		},
	})
}
