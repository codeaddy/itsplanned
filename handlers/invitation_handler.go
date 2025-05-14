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

	inviteLink := fmt.Sprintf("http://localhost:8080/events/redirect/%s", inviteCode)

	c.JSON(http.StatusOK, api.GenerateInviteLinkResponse{InviteLink: inviteLink})
}

// @Summary Redirect to iOS app deeplink
// @Description Redirects from web browser to iOS app deeplink for event joining
// @Tags invitations
// @Produce html
// @Param invite_code path string true "Invite Code"
// @Success 302 {string} string "Redirect to iOS app"
// @Failure 404 {object} api.APIResponse "Invalid invite link"
// @Router /events/redirect/{invite_code} [get]
func DeepLinkRedirect(c *gin.Context, db *gorm.DB) {
	inviteCode := c.Param("invite_code")

	var invitation models.EventInvitation
	if err := db.Where("invite_code = ?", inviteCode).First(&invitation).Error; err != nil {
		c.JSON(http.StatusNotFound, api.APIResponse{Error: "Invalid invite link"})
		return
	}

	deepLink := fmt.Sprintf("itsplanned://event/join?code=%s", inviteCode)

	html := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta http-equiv="refresh" content="0;url=%s">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>Redirecting to ItsPlanned</title>
			<style>
				body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; text-align: center; padding: 20px; }
				.container { max-width: 500px; margin: 50px auto; }
				.btn { display: inline-block; background-color: #4CAF50; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; font-weight: bold; margin-top: 20px; }
			</style>
		</head>
		<body>
			<div class="container">
				<h1>Redirecting to ItsPlanned app...</h1>
				<p>If you're not redirected automatically, click the button below:</p>
				<a href="%s" class="btn">Open in ItsPlanned App</a>
			</div>
		</body>
		</html>
	`, deepLink, deepLink)

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

// @Summary Join event using invite link
// @Description Join an event using a unique invite code
// @Tags invitations
// @Produce json
// @Security BearerAuth
// @Param invite_code path string false "Invite Code from path"
// @Param code query string false "Invite Code from query parameter"
// @Success 200 {object} api.JoinEventResponse "Successfully joined event"
// @Failure 400 {object} api.APIResponse "Already a participant"
// @Failure 401 {object} api.APIResponse "Unauthorized"
// @Failure 404 {object} api.APIResponse "Invalid invite link"
// @Router /events/join/{invite_code} [get]
func JoinEvent(c *gin.Context, db *gorm.DB) {
	// Get invite code from path parameter or query parameter
	inviteCode := c.Param("invite_code")
	if inviteCode == "" {
		// If not in path, try to get from query parameter
		inviteCode = c.Query("code")
	}

	if inviteCode == "" {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Invite code is required"})
		return
	}

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

// @Summary Leave an event
// @Description Allow a user to leave an event they are participating in
// @Tags invitations
// @Produce json
// @Security BearerAuth
// @Param id path int true "Event ID"
// @Success 200 {object} api.APIResponse "Successfully left event"
// @Failure 400 {object} api.APIResponse "Invalid event ID"
// @Failure 401 {object} api.APIResponse "Unauthorized"
// @Failure 403 {object} api.APIResponse "Cannot leave - you are the organizer"
// @Failure 404 {object} api.APIResponse "Not a participant or event not found"
// @Router /events/{id}/leave [delete]
func LeaveEvent(c *gin.Context, db *gorm.DB) {
	eventIDStr := c.Param("id")
	if eventIDStr == "" {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Event ID is required"})
		return
	}

	var eventID uint
	if _, err := fmt.Sscanf(eventIDStr, "%d", &eventID); err != nil {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Invalid event ID format"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, api.APIResponse{Error: "You must be logged in to leave an event"})
		return
	}

	// Check if the event exists
	var event models.Event
	if err := db.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, api.APIResponse{Error: "Event not found"})
		return
	}

	// Prevent organizer from leaving their own event
	if event.OrganizerID == userID.(uint) {
		c.JSON(http.StatusForbidden, api.APIResponse{Error: "Organizers cannot leave their own events"})
		return
	}

	// Check if the user is a participant
	var participation models.EventParticipation
	if err := db.Where("event_id = ? AND user_id = ?", eventID, userID).First(&participation).Error; err != nil {
		c.JSON(http.StatusNotFound, api.APIResponse{Error: "You are not a participant of this event"})
		return
	}

	// Delete the participation record
	if err := db.Delete(&participation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, api.APIResponse{Error: "Failed to leave event"})
		return
	}

	c.JSON(http.StatusOK, api.APIResponse{Message: "Successfully left event"})
}
