package handlers

import (
	"fmt"
	"itsplanned/models"
	"itsplanned/models/api"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// @Summary Generate event invite link
// @Description Generate a unique invite link for an event
// @Tags invitations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body api.GenerateInviteLinkRequest true "Event ID"
// @Success 200 {object} api.GenerateInviteLinkResponse "Invite link generated successfully"
// @Failure 400 {object} api.APIResponse "Invalid payload"
// @Failure 401 {object} api.APIResponse "Unauthorized"
// @Failure 404 {object} api.APIResponse "Event not found"
// @Router /events/invite [post]
func GenerateInviteLink(c *gin.Context, db *gorm.DB) {
	var request api.GenerateInviteLinkRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Invalid payload"})
		return
	}

	var event models.Event
	if err := db.First(&event, request.EventID).Error; err != nil {
		c.JSON(http.StatusNotFound, api.APIResponse{Error: "Event not found"})
		return
	}

	inviteCode := models.GenerateUniqueInviteCode(db)

	invitation := models.EventInvitation{
		EventID:    request.EventID,
		InviteCode: inviteCode,
	}
	db.Create(&invitation)

	inviteLink := fmt.Sprintf("http://localhost:8080/events/join/%s", inviteCode)

	c.JSON(http.StatusOK, api.GenerateInviteLinkResponse{InviteLink: inviteLink})
}

// @Summary Join event using invite link
// @Description Join an event using a unique invite code
// @Tags invitations
// @Produce json
// @Security BearerAuth
// @Param invite_code path string true "Invite Code"
// @Success 200 {object} api.JoinEventResponse "Successfully joined event"
// @Failure 400 {object} api.APIResponse "Already a participant"
// @Failure 401 {object} api.APIResponse "Unauthorized"
// @Failure 404 {object} api.APIResponse "Invalid invite link"
// @Router /events/join/{invite_code} [get]
func JoinEvent(c *gin.Context, db *gorm.DB) {
	inviteCode := c.Param("invite_code")

	var invitation models.EventInvitation
	if err := db.Where("invite_code = ?", inviteCode).First(&invitation).Error; err != nil {
		c.JSON(http.StatusNotFound, api.APIResponse{Error: "Invalid invite link"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, api.APIResponse{Error: "You must be logged in to join"})
		return
	}

	var existingParticipant models.EventParticipation
	if err := db.Where("event_id = ? AND user_id = ?", invitation.EventID, userID).First(&existingParticipant).Error; err == nil {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "You are already in this event"})
		return
	}

	participation := models.EventParticipation{
		EventID: invitation.EventID,
		UserID:  userID.(uint),
	}
	db.Create(&participation)

	c.JSON(http.StatusOK, api.JoinEventResponse{Message: "Successfully joined event"})
}
