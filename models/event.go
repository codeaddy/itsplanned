package models

import (
	"time"

	"gorm.io/gorm"
)

type Event struct {
	gorm.Model
	Name          string    `gorm:"not null"`
	Description   string    `gorm:"type:text"`
	EventDateTime time.Time `gorm:"not null"`
	InitialBudget float64
	OrganizerID   uint
	Place         string       `gorm:"type:text"`
	Tasks         []Task       `gorm:"foreignKey:EventID"`
	EventScores   []EventScore `gorm:"foreignKey:EventID"`
}

type EventScore struct {
	gorm.Model
	EventID uint    `gorm:"not null"`
	UserID  uint    `gorm:"not null"`
	Score   float64 `gorm:"default:0"`
}

type CalendarEvent struct {
	gorm.Model
	UserID    uint      `gorm:"not null"`
	Title     string    `gorm:"not null"`
	StartTime time.Time `gorm:"not null"` // В формате RFC3339
	EndTime   time.Time `gorm:"not null"` // В формате RFC3339
}

func MigrateEvent(db *gorm.DB) error {
	return db.AutoMigrate(&Event{}, &EventScore{}, &CalendarEvent{})
}
