package models

import "gorm.io/gorm"

type UserToken struct {
	ID           uint   `gorm:"primaryKey"`
	UserID       uint   `gorm:"not null"` // Владелец токена
	Provider     string `gorm:"not null"` // google или apple
	AccessToken  string `gorm:"not null"` // Зашифрованный токен
	RefreshToken string `gorm:"not null"` // Зашифрованный refresh-токен
	Expiry       string `gorm:"not null"` // Срок действия токена
}

func MigrateToken(db *gorm.DB) error {
	return db.AutoMigrate(&UserToken{})
}
