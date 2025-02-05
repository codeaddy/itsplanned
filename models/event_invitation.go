package models

import (
	"crypto/rand"
	"encoding/base64"

	"gorm.io/gorm"
)

type EventInvitation struct {
	ID         uint   `gorm:"primaryKey"`
	EventID    uint   `gorm:"not null"`
	InviteCode string `gorm:"unique;not null"`
}

func GenerateUniqueInviteCode(db *gorm.DB) string {
	for {
		b := make([]byte, 12)
		rand.Read(b)
		code := base64.URLEncoding.EncodeToString(b)

		// Проверяем, есть ли сгенерированный код в базе
		var existing EventInvitation
		if err := db.Where("invite_code = ?", code).First(&existing).Error; err != nil {
			return code
		}
	}
}

func MigrateEventInvitation(db *gorm.DB) error {
	return db.AutoMigrate(&EventInvitation{})
}
