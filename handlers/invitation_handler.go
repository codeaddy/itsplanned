package handlers

import (
	"fmt"
	"itsplanned/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GenerateInviteLink(c *gin.Context, db *gorm.DB) {
	var payload struct {
		EventID uint `json:"event_id"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	// TODO Только создатель должен иметь возможность приглашать людей
	var event models.Event
	if err := db.First(&event, payload.EventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	inviteCode := models.GenerateUniqueInviteCode(db)

	invitation := models.EventInvitation{
		EventID:    payload.EventID,
		InviteCode: inviteCode,
	}
	db.Create(&invitation)

	inviteLink := fmt.Sprintf("http://localhost:8080/events/join/%s", inviteCode)

	c.JSON(http.StatusOK, gin.H{"invite_link": inviteLink})
}

func JoinEvent(c *gin.Context, db *gorm.DB) {
	inviteCode := c.Param("invite_code")

	var invitation models.EventInvitation
	if err := db.Where("invite_code = ?", inviteCode).First(&invitation).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid invite link"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You must be logged in to join"})
		return
	}

	var existingParticipant models.EventParticipation
	if err := db.Where("event_id = ? AND user_id = ?", invitation.EventID, userID).First(&existingParticipant).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You are already in this event"})
		return
	}

	participation := models.EventParticipation{
		EventID: invitation.EventID,
		UserID:  userID.(uint),
	}
	db.Create(&participation)

	c.JSON(http.StatusOK, gin.H{"message": "Successfully joined event"})
}
