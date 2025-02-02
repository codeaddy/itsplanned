package models

import "gorm.io/gorm"

type Event struct {
	ID            uint   `gorm:"primaryKey"`
	Name          string `gorm:"not null"`
	Description   string `gorm:"type:text"`
	InitialBudget float64
	OrganizerID   uint
	Tasks         []Task `gorm:"foreignKey:EventID"`
}

func MigrateEvent(db *gorm.DB) error {
	return db.AutoMigrate(&Event{})
}
