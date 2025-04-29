package models

import "gorm.io/gorm"

type AIChat struct {
	gorm.Model
	UserID uint `gorm:"not null"`
}

type AIMessage struct {
	gorm.Model
	ChatID  uint   `gorm:"not null"`
	UserID  uint   `gorm:"not null"`
	Content string `gorm:"type:text;not null"`
	IsUser  bool   `gorm:"not null"`
}

func MigrateAIChat(db *gorm.DB) error {
	return db.AutoMigrate(&AIChat{}, &AIMessage{})
}
