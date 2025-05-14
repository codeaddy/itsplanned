package main

import (
	"itsplanned/common"
	"itsplanned/models"
	"itsplanned/routes"
	"itsplanned/services/email"
	"itsplanned/services/scheduler"
	"itsplanned/services/yandex"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "itsplanned/docs"

	"github.com/joho/godotenv"
)

// @title           ItsPlanned API
// @version         1.0
// @description     API Server for ItsPlanned - A Collaborative Event Planning Application. This API provides endpoints for managing events, tasks, user profiles, and integrations with external services like Google Calendar.

// @contact.name   ItsPlanned Support
// @contact.url    https://github.com/vl4ddos/itsplanned
// @contact.email  support@itsplanned.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter the token with the `Bearer: ` prefix, e.g. "Bearer abcde12345"

// @tag.name auth
// @tag.description Authentication endpoints for user registration, login, and password management

// @tag.name profile
// @tag.description User profile management endpoints

// @tag.name events
// @tag.description Event management endpoints for creating, updating, and managing events

// @tag.name tasks
// @tag.description Task management endpoints for creating, assigning, and completing tasks

// @tag.name invitations
// @tag.description Event invitation endpoints for generating and using invite links

// @tag.name ai-assistant
// @tag.description AI assistant endpoints for chat-based event planning assistance

// @tag.name calendar
// @tag.description Google Calendar integration endpoints for importing events

// @tag.name notifications
// @tag.description Push notification endpoints for registering device tokens and managing notification preferences
func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		log.Fatal("DATABASE_DSN environment variable is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	app := &common.App{DB: db}

	if err := email.Init(); err != nil {
		log.Fatal("Error initializing email service:", err)
	}

	if err := yandex.Init(); err != nil {
		log.Fatal("Error initializing Yandex token service:", err)
	}

	taskScheduler := scheduler.NewScheduler(db)
	taskScheduler.SetupCalendarSyncTask()
	taskScheduler.Start()
	log.Println("Started calendar sync background task")

	// Run database migrations
	if err := models.MigrateUser(db); err != nil {
		log.Fatal("Failed to migrate user model: ", err)
	}
	if err := models.MigrateEvent(db); err != nil {
		log.Fatal("Failed to migrate event model: ", err)
	}
	if err := models.MigrateTask(db); err != nil {
		log.Fatal("Failed to migrate task model: ", err)
	}
	if err := models.MigrateToken(db); err != nil {
		log.Fatal("Failed to migrate user token model: ", err)
	}
	if err := models.MigrateEventInvitation(db); err != nil {
		log.Fatal("Failed to migrate event invitation model: ", err)
	}
	if err := models.MigrateEventParticipation(db); err != nil {
		log.Fatal("Failed to migrate event participation model: ", err)
	}
	if err := models.MigratePasswordReset(db); err != nil {
		log.Fatal("Failed to migrate password reset model: ", err)
	}
	if err := models.MigrateTaskStatusEvent(db); err != nil {
		log.Fatal("Failed to migrate task status event model: ", err)
	}
	if err := models.MigrateAIChat(db); err != nil {
		log.Fatal("Failed to migrate AI chat model: ", err)
	}

	// Setup routes
	r := routes.SetupRouter(app)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server is running on http://localhost:" + port)
	log.Println("Swagger documentation is available at http://localhost:" + port + "/swagger/index.html")
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Error starting server:", err)
	}
}
