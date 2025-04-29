package models

import (
	"time"

	"gorm.io/gorm"
)

type TaskStatusEvent struct {
	gorm.Model
	TaskID        uint      `gorm:"not null;index"`
	TaskName      string    `gorm:"type:varchar(255);not null"`
	OldStatus     string    `gorm:"type:varchar(20)"`
	NewStatus     string    `gorm:"type:varchar(20);not null"`
	UserID        uint      `gorm:"not null;index"`
	ChangedByID   uint      `gorm:"not null"`
	ChangedByName string    `gorm:"type:varchar(255);not null"`
	IsRead        bool      `gorm:"default:false"`
	EventTime     time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func MigrateTaskStatusEvent(db *gorm.DB) error {
	return db.AutoMigrate(&TaskStatusEvent{})
}
