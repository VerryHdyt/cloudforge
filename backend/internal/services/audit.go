package services

import (
	"github.com/VerryHdyt/cloudforge/backend/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuditService struct {
	DB *gorm.DB
}

type AuditEntry struct {
	UserID     *uint
	Action     string
	Resource   string
	ResourceID string
	IPAddress  string
	UserAgent  string
}

func (s *AuditService) Log(entry AuditEntry) error {
	auditLog := models.AuditLog{
		UserID:     entry.UserID,
		Action:     entry.Action,
		Resource:   entry.Resource,
		ResourceID: entry.ResourceID,
		IPAddress:  entry.IPAddress,
		UserAgent:  entry.UserAgent,
	}

	return s.DB.Create(&auditLog).Error
}

func (s *AuditService) LogFromContext(
	c *gin.Context,
	userID *uint,
	action string,
	resource string,
	resourceID string,
) error {
	return s.Log(AuditEntry{
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})
}
