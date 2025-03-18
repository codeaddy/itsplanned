package test

import (
	"fmt"
	"itsplanned/common"
	"itsplanned/models"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var TestDB *gorm.DB
var TestApp *common.App

// SetupTestDB initializes a test database and returns a cleanup function
func SetupTestDB(t *testing.T) func() {
	// Use SQLite for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations
	err = db.AutoMigrate(
		&models.User{},
		&models.Event{},
		&models.Task{},
		&models.EventParticipation{},
		&models.EventScore{},
		&models.CalendarEvent{},
		&models.AIChat{},
		&models.AIMessage{},
	)
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	TestDB = db
	TestApp = &common.App{
		DB: db,
	}

	// Return cleanup function
	return func() {
		sqlDB, err := db.DB()
		if err != nil {
			t.Logf("Failed to get database instance: %v", err)
			return
		}
		sqlDB.Close()
	}
}

// CreateTestUser creates a test user and returns it
func CreateTestUser(t *testing.T) *models.User {
	user := &models.User{
		Email:       fmt.Sprintf("test%d@example.com", time.Now().UnixNano()),
		DisplayName: "Test User",
	}

	if err := TestDB.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return user
}

// CreateTestEvent creates a test event and returns it
func CreateTestEvent(t *testing.T, organizerID uint) *models.Event {
	eventTime, _ := time.Parse(time.RFC3339, "2024-04-01T18:00:00Z")
	event := &models.Event{
		Name:          "Test Event",
		Description:   "Test Description",
		EventDateTime: eventTime,
		InitialBudget: 1000.0,
		OrganizerID:   organizerID,
		Place:         "Test Place",
	}

	if err := TestDB.Create(event).Error; err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	// Create event participation for the organizer
	participation := &models.EventParticipation{
		EventID: event.ID,
		UserID:  organizerID,
	}

	if err := TestDB.Create(participation).Error; err != nil {
		t.Fatalf("Failed to create event participation: %v", err)
	}

	return event
}

// CreateTestTask creates a test task and returns it
func CreateTestTask(t *testing.T, eventID uint) *models.Task {
	task := &models.Task{
		Title:       "Test Task",
		Description: "Test Description",
		Budget:      100.0,
		Points:      10,
		EventID:     eventID,
	}

	if err := TestDB.Create(task).Error; err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}

	return task
}

// CreateTestContext creates a Gin context for testing
func CreateTestContext(t *testing.T, userID uint) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", userID)
	return c, w
}

// AddEventParticipant adds a user as a participant to an event
func AddEventParticipant(t *testing.T, eventID, userID uint) {
	participation := &models.EventParticipation{
		EventID: eventID,
		UserID:  userID,
	}

	if err := TestDB.Create(participation).Error; err != nil {
		t.Fatalf("Failed to add event participant: %v", err)
	}
}
