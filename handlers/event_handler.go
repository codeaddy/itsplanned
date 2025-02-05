package handlers

import (
	"fmt"
	"itsplanned/models"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const leftTimeBoundForBusySlots string = "08:00"
const rightTimeBoundForBusySlots string = "22:00"

func getUserBusySlotsForDay(db *gorm.DB, userID uint, date string, busySlots *map[string]int) {
	var events []models.CalendarEvent
	// dateCasted, _ := time.Parse("2006-01-02", date)
	db.Where("user_id = ? AND DATE(start_time) = ?", userID, date).Find(&events)

	fmt.Println(len(events))

	for _, event := range events {
		start := event.StartTime
		end := event.EndTime

		for start.Minute() != 0 && start.Minute() != 30 {
			start = start.Add(-time.Minute)
		}

		for end.Minute() != 0 && end.Minute() != 30 {
			end = end.Add(time.Minute)
		}

		for t := start; t.Before(end); t = t.Add(time.Minute * 30) {
			key := t.Format("15:04")
			(*busySlots)[key]++
			fmt.Println(key)
		}
	}
}

func suggestTimeSlotsForDay(busySlots *map[string]int, date string, durationMins int64) []map[string]interface{} {
	start, _ := time.Parse("15:04", leftTimeBoundForBusySlots)
	end, _ := time.Parse("15:04", rightTimeBoundForBusySlots)

	var timeSlots []map[string]interface{}
	timeCursor := start

	for timeCursor.Before(end) {
		maxBusy := 0

		for i := int64(0); i < durationMins; i += 30 {
			key := timeCursor.Add(time.Minute * time.Duration(i)).Format("15:04")
			if (*busySlots)[key] > 0 {
				if (*busySlots)[key] > maxBusy {
					maxBusy = (*busySlots)[key]
				}
			}
		}

		timeSlots = append(timeSlots, map[string]interface{}{
			"slot":       date + " " + timeCursor.Format("15:04"),
			"busy_count": maxBusy,
		})

		timeCursor = timeCursor.Add(time.Minute * 30)
	}

	// Сортируем по возрастанию количества занятых участников (чем меньше, тем лучше)
	sort.Slice(timeSlots, func(i, j int) bool {
		return timeSlots[i]["busy_count"].(int) < timeSlots[j]["busy_count"].(int)
	})

	// Возвращаем 5 лучших вариантов
	if len(timeSlots) > 5 {
		timeSlots = timeSlots[:5]
	}

	return timeSlots
}

func CreateEvent(c *gin.Context, db *gorm.DB) {
	var payload struct {
		Name          string  `json:"name"`
		Description   string  `json:"description"`
		EventDateTime string  `json:"event_date_time"`
		InitialBudget float64 `json:"initial_budget"` // начальный бюджет
		OrganizerID   uint    `json:"organizer_id"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	eventTime, err := time.Parse(time.RFC3339, payload.EventDateTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use RFC3339 format."})
		return
	}

	event := models.Event{
		Name:          payload.Name,
		Description:   payload.Description,
		EventDateTime: eventTime,
		InitialBudget: payload.InitialBudget,
		OrganizerID:   payload.OrganizerID,
	}

	if err := db.Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event"})
		return
	}

	eventParticipation := models.EventParticipation{
		EventID: event.ID,
		UserID:  event.OrganizerID,
	}

	if err := db.Create(&eventParticipation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event_participation"})
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
		Name          *string `json:"name"`
		Description   *string `json:"description"`
		EventDateTime *string `json:"event_date_time"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	if payload.Name != nil {
		event.Name = *payload.Name
	}
	if payload.Description != nil {
		event.Description = *payload.Description
	}
	if payload.EventDateTime != nil {
		eventTime, err := time.Parse(time.RFC3339, *payload.EventDateTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
			return
		}
		event.EventDateTime = eventTime
	}

	db.Save(&event)
	c.JSON(http.StatusOK, gin.H{"message": "Event updated successfully", "event": event})
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

func GetEventLeaderboard(c *gin.Context, db *gorm.DB) {
	userID, _ := c.Get("user_id")
	eventID := c.Param("id")

	var participant models.EventScore
	if err := db.Where("event_id = ? AND user_id = ?", eventID, userID).First(&participant).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not a participant of this event"})
		return
	}

	var leaderboard []models.EventScore
	if err := db.Where("event_id = ?", eventID).Order("score DESC").Find(&leaderboard).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Leaderboard not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"leaderboard": leaderboard})
}

func GetEvents(c *gin.Context, db *gorm.DB) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var events []models.Event
	db.Joins("JOIN event_participations ON events.id = event_participations.event_id").
		Where("event_participations.user_id = ?", userID).
		Find(&events)

	c.JSON(http.StatusOK, gin.H{"events": events})
}

func FindBestTimeSlotsForDay(c *gin.Context, db *gorm.DB) {
	var payload struct {
		EventID      uint   `json:"event_id"`
		DurationMins int64  `json:"duration_mins"`
		Date         string `json:"date"`
	}

	fmt.Println("WE ENTERED HERE 1")

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	var event models.Event
	if err := db.First(&event, payload.EventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	var participants []models.User
	db.Joins("JOIN event_participations ON event_participations.user_id = users.id").
		Where("event_participations.event_id = ?", payload.EventID).
		Find(&participants)

	if len(participants) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No participants found"})
		return
	}

	fmt.Println("WE ENTERED HERE 2")

	busySlots := make(map[string]int)
	for _, user := range participants {
		getUserBusySlotsForDay(db, user.ID, payload.Date, &busySlots)
	}

	fmt.Println(busySlots)

	suggestedSlots := suggestTimeSlotsForDay(&busySlots, payload.Date, payload.DurationMins)

	c.JSON(http.StatusOK, gin.H{"suggested_slots": suggestedSlots})
}
