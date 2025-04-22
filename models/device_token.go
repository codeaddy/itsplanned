package models

import "gorm.io/gorm"

// DeviceToken represents a user's device token for push notifications
type DeviceToken struct {
	ID         uint   `gorm:"primaryKey"`
	UserID     uint   `gorm:"not null;index"`
	Token      string `gorm:"not null;index"`
	DeviceType string `gorm:"not null"` // ios, android, etc.
	IsActive   bool   `gorm:"default:true"`
	CreatedAt  int64  `gorm:"autoCreateTime"`
	LastUsedAt int64  `gorm:"autoUpdateTime"`
}

// MigrateDeviceToken creates or updates the device token table
func MigrateDeviceToken(db *gorm.DB) error {
	return db.AutoMigrate(&DeviceToken{})
}
