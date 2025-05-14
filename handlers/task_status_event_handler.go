package handlers

import (
	"itsplanned/models"
	"itsplanned/models/api"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func toTaskStatusEventResponse(event *models.TaskStatusEvent) api.TaskStatusEventResponse {
	return api.TaskStatusEventResponse{
		ID:            event.ID,
		TaskID:        event.TaskID,
		TaskName:      event.TaskName,
		OldStatus:     event.OldStatus,
		NewStatus:     event.NewStatus,
		ChangedByID:   event.ChangedByID,
		ChangedByName: event.ChangedByName,
		IsRead:        event.IsRead,
		EventTime:     event.EventTime,
	}
}

// @Summary Get unread task status events
// @Description Get all unread task status events for the authenticated user
// @Tags events
// @Produce json
// @Security BearerAuth
// @Success 200 {object} api.TaskStatusEventsResponse "Unread task status events retrieved successfully"
// @Failure 401 {object} api.APIResponse "Unauthorized"
// @Failure 500 {object} api.APIResponse "Internal server error"
// @Router /task-status-events/unread [get]
func GetUnreadTaskStatusEvents(c *gin.Context, db *gorm.DB) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, api.APIResponse{Error: "User not authenticated"})
		return
	}

	var events []models.TaskStatusEvent
	if err := db.Where("user_id = ? AND is_read = ?", userID, false).
		Order("event_time DESC").
		Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, api.APIResponse{Error: "Failed to retrieve task status events"})
		return
	}

	var responseEvents []api.TaskStatusEventResponse
	for _, event := range events {
		responseEvents = append(responseEvents, toTaskStatusEventResponse(&event))
	}

	if len(events) > 0 {
		if err := db.Model(&models.TaskStatusEvent{}).
			Where("user_id = ? AND is_read = ?", userID, false).
			Update("is_read", true).Error; err != nil {

			c.JSON(http.StatusInternalServerError, api.APIResponse{Error: "Failed to mark events as read"})
			return
		}
	}

	c.JSON(http.StatusOK, api.TaskStatusEventsResponse{Events: responseEvents})
}
