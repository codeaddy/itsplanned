package models

import "gorm.io/gorm"

type EventParticipation struct {
	ID      uint `gorm:"primaryKey"`
	EventID uint `gorm:"not null;index"`
	UserID  uint `gorm:"not null;index"`
}

func MigrateEventParticipation(db *gorm.DB) error {
	return db.AutoMigrate(&EventParticipation{})
}
