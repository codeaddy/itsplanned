package models

import (
	"time"

	"gorm.io/gorm"
)

type PasswordReset struct {
	gorm.Model
	UserID     uint
	Token      string
	ExpiryTime time.Time
	Used       bool
	User       User `gorm:"foreignKey:UserID"`
}

func MigratePasswordReset(db *gorm.DB) error {
	return db.AutoMigrate(&PasswordReset{})
}
