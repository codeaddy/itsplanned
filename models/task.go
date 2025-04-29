package models

import "gorm.io/gorm"

type Task struct {
	ID          uint    `gorm:"primaryKey"`
	Title       string  `gorm:"not null"`
	Description string  `gorm:"default:''"`
	IsCompleted bool    `gorm:"default:false"`
	Budget      float64 `gorm:"not null"`
	Points      int     `gorm:"not null"`
	EventID     uint    `gorm:"not null"`
	AssignedTo  *uint   `gorm:"default:null"`
}

func MigrateTask(db *gorm.DB) error {
	return db.AutoMigrate(&Task{})
}
