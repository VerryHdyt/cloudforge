package handlers

import (
	"net/http"
	"strconv"

	"github.com/VerryHdyt/cloudforge/backend/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuditHandler struct {
	DB *gorm.DB
}

func (h *AuditHandler) GetAuditLogs(c *gin.Context) {
	// Pagination
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit < 1 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	// Query database
	query := h.DB.Model(&models.AuditLog{})

	// Filter by action
	if action := c.Query("action"); action != "" {
		query = query.Where("action = ?", action)
	}

	// Filter by resource
	if resource := c.Query("resource"); resource != "" {
		query = query.Where("resource = ?", resource)
	}

	// Count total records
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to count audit logs",
		})
		return
	}

	// Retrieve audit logs
	var logs []models.AuditLog

	if err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve audit logs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": logs,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}
