package handlers

import (
	"context"
	"fmt"
	"itsplanned/models"
	"itsplanned/models/api"
	"itsplanned/security"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
	"gorm.io/gorm"
)

var (
	// For testing purposes
	newCalendarService = calendar.NewService
)

func getGoogleClient(ctx context.Context, accessToken string) *http.Client {
	config := &oauth2.Config{
		Endpoint: google.Endpoint,
	}

	token := &oauth2.Token{AccessToken: accessToken}
	return config.Client(ctx, token)
}

func ImportCalendarEventsForUser(db *gorm.DB, userID uint, accessToken string) error {
	ctx := context.Background()
	client := getGoogleClient(ctx, accessToken)
	srv, err := newCalendarService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Printf("Unable to retrieve Calendar client: %v", err)
		return fmt.Errorf("failed to connect to Google Calendar: %v", err)
	}

	now := time.Now().Format(time.RFC3339)
	fourWeeksLater := time.Now().AddDate(0, 0, 28).Format(time.RFC3339)

	events, err := srv.Events.List("primary").
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(now).
		TimeMax(fourWeeksLater).
		OrderBy("startTime").
		Do()
	if err != nil {
		log.Printf("Unable to retrieve events: %v", err)
		return fmt.Errorf("unable to retrieve events: %v", err)
	}

	for _, item := range events.Items {
		if item.Start.DateTime == "" || item.End.DateTime == "" {
			continue
		}

		startTime, err := time.Parse(time.RFC3339, item.Start.DateTime)
		if err != nil {
			log.Printf("Error parsing start time for event %s: %v", item.Id, err)
			continue
		}
		startTime = startTime.In(time.FixedZone("MSK", 3*60*60))

		endTime, err := time.Parse(time.RFC3339, item.End.DateTime)
		if err != nil {
			log.Printf("Error parsing end time for event %s: %v", item.Id, err)
			continue
		}
		endTime = endTime.In(time.FixedZone("MSK", 3*60*60))

		var existingEvent models.CalendarEvent
		if err := db.Where("user_id = ? AND title = ? AND start_time = ?", userID, item.Summary, startTime).First(&existingEvent).Error; err == nil {
			db.Model(&existingEvent).Updates(models.CalendarEvent{
				Title:   item.Summary,
				EndTime: endTime,
			})
			continue
		}

		event := models.CalendarEvent{
			UserID:    userID,
			Title:     item.Summary,
			StartTime: startTime,
			EndTime:   endTime,
		}

		if err := db.Create(&event).Error; err != nil {
			log.Printf("Error creating calendar event for user %d: %v", userID, err)
		}
	}

	return nil
}

// @Summary Import Google Calendar events
// @Description Import events from the user's Google Calendar for the next 4 weeks
// @Tags calendar
// @Produce json
// @Security BearerAuth
// @Success 200 {object} api.ImportCalendarEventsResponse "Events imported successfully"
// @Failure 401 {object} api.APIResponse "Unauthorized or no token found"
// @Failure 500 {object} api.APIResponse "Failed to import events"
// @Router /calendar/import [get]
func ImportCalendarEvents(c *gin.Context, db *gorm.DB) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, api.APIResponse{Error: "Unauthorized"})
		return
	}

	var token models.UserToken
	if err := db.Where("user_id = ?", userID).First(&token).Error; err != nil {
		c.JSON(http.StatusUnauthorized, api.APIResponse{Error: "No token found"})
		return
	}

	decryptedAccessToken, err := security.DecryptToken(token.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.APIResponse{Error: "Failed to decrypt token"})
		return
	}

	err = ImportCalendarEventsForUser(db, userID.(uint), decryptedAccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.APIResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, api.ImportCalendarEventsResponse{Message: "Events imported successfully"})
}
