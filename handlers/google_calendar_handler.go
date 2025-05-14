package handlers

import (
	"itsplanned/calendar"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func FetchGoogleCalendarEventsHandler(c *gin.Context) {
	var payload struct {
		AccessToken string `json:"access_token"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	startDate, err := time.Parse(time.RFC3339, payload.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format"})
		return
	}

	endDate, err := time.Parse(time.RFC3339, payload.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format"})
		return
	}

	events, err := calendar.FetchGoogleCalendarEvents(payload.AccessToken, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"events": events})
}
