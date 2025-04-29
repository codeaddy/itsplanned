package models

import "gorm.io/gorm"

type UserToken struct {
	ID           uint   `gorm:"primaryKey"`
	UserID       uint   `gorm:"not null"`
	AccessToken  string `gorm:"not null"`
	RefreshToken string `gorm:"not null"`
	Expiry       string `gorm:"not null"`
}

func MigrateToken(db *gorm.DB) error {
	return db.AutoMigrate(&UserToken{})
}
