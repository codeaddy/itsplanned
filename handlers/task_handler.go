package handlers

import (
	"itsplanned/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateTask(c *gin.Context, db *gorm.DB) {
	var payload struct {
		Title      string  `json:"title"`
		Budget     float64 `json:"budget"`
		Points     int     `json:"points"`
		EventID    uint    `json:"event_id"`
		AssignedTo *uint   `json:"assigned_to"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	task := models.Task{
		Title:      payload.Title,
		Budget:     payload.Budget,
		Points:     payload.Points,
		EventID:    payload.EventID,
		AssignedTo: payload.AssignedTo,
	}

	if err := db.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task created", "task": task})
}

func AssignToTask(c *gin.Context, db *gorm.DB) {
	var task models.Task
	id := c.Param("id")

	if err := db.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if task.AssignedTo != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task already has an assignee"})
		return
	}

	var payload struct {
		UserID uint `json:"user_id"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	task.AssignedTo = &payload.UserID
	db.Save(&task)

	c.JSON(http.StatusOK, gin.H{"message": "You have been assigned to the task", "task": task})
}

func CompleteTask(c *gin.Context, db *gorm.DB) {
	var task models.Task
	id := c.Param("id")

	if err := db.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if task.IsCompleted {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task already completed"})
		return
	}

	userID, _ := c.Get("user_id")
	if task.AssignedTo == nil || userID.(uint) != *task.AssignedTo {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only complete your own tasks"})
		return
	}

	task.IsCompleted = true
	db.Save(&task)

	var user models.User
	if err := db.First(&user, task.AssignedTo).Error; err == nil {
		user.TotalScore += task.Points
		db.Save(&user)
	}

	var eventScore models.EventScore
	if err := db.Where("event_id = ? AND user_id = ?", task.EventID, task.AssignedTo).First(&eventScore).Error; err != nil {
		eventScore = models.EventScore{
			EventID: task.EventID,
			UserID:  *task.AssignedTo,
			Score:   task.Points,
		}
		db.Create(&eventScore)
	} else {
		eventScore.Score += task.Points
		db.Save(&eventScore)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task completed", "task": task})
}
