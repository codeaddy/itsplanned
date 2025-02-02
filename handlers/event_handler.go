package handlers

import (
	"itsplanned/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateEvent(c *gin.Context, db *gorm.DB) {
	var payload struct {
		Name          string  `json:"name"`
		Description   string  `json:"description"`
		InitialBudget float64 `json:"initial_budget"` // начальный бюджет
		OrganizerID   uint    `json:"organizer_id"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	event := models.Event{
		Name:          payload.Name,
		Description:   payload.Description,
		InitialBudget: payload.InitialBudget,
		OrganizerID:   payload.OrganizerID,
	}

	if err := db.Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event created", "event": event})
}

func UpdateEvent(c *gin.Context, db *gorm.DB) {
	eventID := c.Param("id")

	var event models.Event
	if err := db.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	userID, _ := c.Get("user_id")
	if userID.(uint) != event.OrganizerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not the organizer of this event"})
		return
	}

	var payload struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	event.Name = payload.Name
	event.Description = payload.Description
	db.Save(&event)

	c.JSON(http.StatusOK, gin.H{"event": event})
}

func UpdateEventBudget(c *gin.Context, db *gorm.DB) {
	var event models.Event
	id := c.Param("id")

	if err := db.First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	var payload struct {
		InitialBudget float64 `json:"initial_budget"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	event.InitialBudget = payload.InitialBudget
	db.Save(&event)

	c.JSON(http.StatusOK, gin.H{"message": "Event budget updated", "event": event})
}

func GetEventBudget(c *gin.Context, db *gorm.DB) {
	var event models.Event
	id := c.Param("id")

	if err := db.Preload("Tasks").First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	realBudget := 0.0
	for _, task := range event.Tasks {
		if task.IsCompleted {
			realBudget += task.Budget
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"initial_budget": event.InitialBudget,
		"real_budget":    realBudget,
		"difference":     event.InitialBudget - realBudget,
	})
}
