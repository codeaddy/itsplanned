package tests

import (
	"fmt"
	"itsplanned/models"
	"log"
	"os"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var testDB *gorm.DB

func SetupTestDB() *gorm.DB {
	dsn := "host=localhost user=postgres password=test dbname=testdb port=5433 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to test DB: %v", err)
	}

	err = db.AutoMigrate(&models.User{}, &models.Event{}, &models.Task{}, &models.EventScore{})
	if err != nil {
		log.Fatalf("Failed to migrate test DB: %v", err)
	}

	return db
}

func TearDownDB(db *gorm.DB) {
	db.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
}

func TestMain(m *testing.M) {
	fmt.Println("Setting up test database...")
	testDB = SetupTestDB()

	code := m.Run()

	fmt.Println("Cleaning up test database...")
	TearDownDB(testDB)

	os.Exit(code)
}
