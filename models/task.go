package models

import "gorm.io/gorm"

type Task struct {
	ID          uint    `gorm:"primaryKey"`
	Title       string  `gorm:"not null"`
	IsCompleted bool    `gorm:"default:false"`
	Budget      float64 `gorm:"not null"` // Начальный бюджет
	Points      int     `gorm:"not null"` // Баллы за задачу
	EventID     uint    `gorm:"not null"`
	AssignedTo  *uint   `gorm:"default:null"` // Исполнитель задачи
}

func MigrateTask(db *gorm.DB) error {
	return db.AutoMigrate(&Task{})
}
