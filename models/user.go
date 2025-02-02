package models

import "gorm.io/gorm"

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Email        string `gorm:"unique;not null"`
	PasswordHash string `gorm:"not null"`
	DisplayName  string `gorm:"not null"`
	TotalScore   int    `gorm:"default:0"`
}

func MigrateUser(db *gorm.DB) error {
	return db.AutoMigrate(&User{})
}
