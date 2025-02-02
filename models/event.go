package models

import "gorm.io/gorm"

type Event struct {
	ID            uint   `gorm:"primaryKey"`
	Name          string `gorm:"not null"`
	Description   string `gorm:"type:text"`
	InitialBudget float64
	OrganizerID   uint
	Tasks         []Task       `gorm:"foreignKey:EventID"`
	EventScores   []EventScore `gorm:"foreignKey:EventID"`
}

type EventScore struct {
	ID      uint `gorm:"primaryKey"`
	EventID uint `gorm:"not null"`
	UserID  uint `gorm:"not null"`
	Score   int  `gorm:"default:0"`
}

func MigrateEvent(db *gorm.DB) error {
	return db.AutoMigrate(&Event{}, &EventScore{})
}
