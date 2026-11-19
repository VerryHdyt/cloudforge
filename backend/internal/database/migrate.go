package database

import (
	"log"

	"github.com/VerryHdyt/cloudforge/backend/internal/models"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	log.Println("Running database migrations...")

	if err := db.AutoMigrate(
		&models.User{},
		&models.AuditLog{},
	); err != nil {
		return err
	}

	log.Println("Database migrations completed successfully")
	return nil
}
