package main

import (
	"itsplanned/common"
	"itsplanned/models"
	"itsplanned/routes"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		log.Fatal("DATABASE_DSN environment variable is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

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

	app := &common.App{DB: db}

	r := routes.SetupRouter(app)

	log.Println("Server is running on http://localhost:8080")
	r.Run(":8080")
}
