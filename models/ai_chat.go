package models

import "gorm.io/gorm"

type AIChat struct {
	gorm.Model
	UserID   uint
	User     User        `gorm:"foreignKey:UserID"`
	Messages []AIMessage `gorm:"foreignKey:ChatID"`
}

type AIMessage struct {
	gorm.Model
	ChatID  uint
	UserID  uint
	Content string
	IsUser  bool
	Chat    AIChat `gorm:"foreignKey:ChatID"`
	User    User   `gorm:"foreignKey:UserID"`
}

func MigrateAIChat(db *gorm.DB) error {
	return db.AutoMigrate(&AIChat{}, &AIMessage{})
}
