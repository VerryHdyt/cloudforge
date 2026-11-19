package models

import "time"

type AuditLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     *uint     `gorm:"index" json:"user_id"`
	Action     string    `gorm:"size:50;not null;index" json:"action"`
	Resource   string    `gorm:"size:100;not null;index" json:"resource"`
	ResourceID string    `gorm:"size:100" json:"resource_id"`
	IPAddress  string    `gorm:"size:45" json:"ip_address"`
	UserAgent  string    `gorm:"type:text" json:"user_agent"`
	CreatedAt  time.Time `gorm:"index" json:"created_at"`
}
