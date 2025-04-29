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

// Creates HTTP-client with OAuth-token
func getGoogleClient(ctx context.Context, accessToken string) *http.Client {
	config := &oauth2.Config{
		Endpoint: google.Endpoint,
	}

	token := &oauth2.Token{AccessToken: accessToken}
	return config.Client(ctx, token)
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

	ctx := context.Background()
	client := getGoogleClient(ctx, decryptedAccessToken)
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Printf("Unable to retrieve Calendar client: %v", err)
		c.JSON(http.StatusInternalServerError, api.APIResponse{Error: "Failed to connect to Google Calendar"})
		return
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
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, api.APIResponse{Error: "Unable to retrieve events"})
		return
	}

	for _, item := range events.Items {
		startTime, err := time.Parse(time.RFC3339, item.Start.DateTime)
		if err != nil {
			continue
		}
		startTime = startTime.In(time.FixedZone("MSK", 3*60*60))

		endTime, err := time.Parse(time.RFC3339, item.End.DateTime)
		if err != nil {
			continue
		}
		endTime = endTime.In(time.FixedZone("MSK", 3*60*60))

		// Check if event already exists (with given title and start_time)
		var existingEvent models.CalendarEvent
		if err := db.Where("user_id = ? AND title = ? AND start_time = ?", userID, item.Summary, startTime).First(&existingEvent).Error; err == nil {
			continue
		}

		event := models.CalendarEvent{
			UserID:    userID.(uint),
			Title:     item.Summary,
			StartTime: startTime,
			EndTime:   endTime,
		}

		db.Create(&event)
	}

	c.JSON(http.StatusOK, api.ImportCalendarEventsResponse{Message: "Events imported successfully"})
}
