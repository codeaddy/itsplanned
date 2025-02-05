package models

import (
	"time"

	"gorm.io/gorm"
)

type Event struct {
	ID            uint      `gorm:"primaryKey"`
	Name          string    `gorm:"not null"`
	Description   string    `gorm:"type:text"`
	EventDateTime time.Time `gorm:"not null"`
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

type CalendarEvent struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null"`
	Title     string    `gorm:"not null"`
	StartTime time.Time `gorm:"not null"` // В формате RFC3339
	EndTime   time.Time `gorm:"not null"` // В формате RFC3339
}

func MigrateEvent(db *gorm.DB) error {
	return db.AutoMigrate(&Event{}, &EventScore{}, &CalendarEvent{})
}
