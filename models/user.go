package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Email        string `gorm:"unique"`
	PasswordHash string
	DisplayName  string
	Bio          string
	Avatar       string
	TotalScore   int
	Events       []Event `gorm:"foreignKey:OrganizerID"`
	Tasks        []Task  `gorm:"foreignKey:AssignedTo"`
}

func MigrateUser(db *gorm.DB) error {
	return db.AutoMigrate(&User{})
}
