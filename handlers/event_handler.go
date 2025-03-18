package handlers

import (
	"fmt"
	"itsplanned/models"
	"itsplanned/models/api"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const leftTimeBoundForBusySlots string = "08:00"
const rightTimeBoundForBusySlots string = "22:00"

func toEventResponse(event *models.Event) *api.EventResponse {
	if event == nil {
		return nil
	}
	return &api.EventResponse{
		ID:            event.ID,
		CreatedAt:     event.CreatedAt,
		UpdatedAt:     event.UpdatedAt,
		Name:          event.Name,
		Description:   event.Description,
		EventDateTime: event.EventDateTime,
		InitialBudget: event.InitialBudget,
		OrganizerID:   event.OrganizerID,
		Place:         event.Place,
	}
}

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

func suggestTimeSlotsForDay(busySlots *map[string]int, date string, durationMins int64, startTime string, endTime string) []api.TimeSlotSuggestion {
	start, _ := time.Parse("15:04", startTime)
	end, _ := time.Parse("15:04", endTime)

	var timeSlots []api.TimeSlotSuggestion
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

		timeSlots = append(timeSlots, api.TimeSlotSuggestion{
			Slot:      date + " " + timeCursor.Format("15:04"),
			BusyCount: maxBusy,
		})

		timeCursor = timeCursor.Add(time.Minute * 30)
	}

	// Sort by ascending number of busy participants (fewer is better)
	sort.Slice(timeSlots, func(i, j int) bool {
		return timeSlots[i].BusyCount < timeSlots[j].BusyCount
	})

	// Return top 5 suggestions
	if len(timeSlots) > 5 {
		timeSlots = timeSlots[:5]
	}

	return timeSlots
}

// @Summary Create a new event
// @Description Create a new event with the given details
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body api.CreateEventRequest true "Event creation details"
// @Success 200 {object} api.APIResponse{data=api.EventResponse} "Event created successfully"
// @Failure 400 {object} api.APIResponse "Invalid payload"
// @Failure 401 {object} api.APIResponse "Unauthorized"
// @Failure 500 {object} api.APIResponse "Failed to create event"
// @Router /events [post]
func CreateEvent(c *gin.Context, db *gorm.DB) {
	var request api.CreateEventRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Invalid payload"})
		return
	}

	eventTime, err := time.Parse(time.RFC3339, request.EventDateTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Invalid date format. Use RFC3339 format."})
		return
	}

	userID, _ := c.Get("user_id")
	event := models.Event{
		Name:          request.Name,
		Description:   request.Description,
		EventDateTime: eventTime,
		InitialBudget: request.InitialBudget,
		OrganizerID:   userID.(uint),
		Place:         request.Place,
	}

	if err := db.Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, api.APIResponse{Error: "Failed to create event"})
		return
	}

	eventParticipation := models.EventParticipation{
		EventID: event.ID,
		UserID:  event.OrganizerID,
	}

	if err := db.Create(&eventParticipation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, api.APIResponse{Error: "Failed to create event_participation"})
		return
	}

	c.JSON(http.StatusOK, api.APIResponse{
		Message: "Event created",
		Data:    toEventResponse(&event),
	})
}

// @Summary Update an event
// @Description Update an existing event's details
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Event ID"
// @Param request body api.UpdateEventRequest true "Event update details"
// @Success 200 {object} api.APIResponse{data=api.EventResponse} "Event updated successfully"
// @Failure 400 {object} api.APIResponse "Invalid payload"
// @Failure 401 {object} api.APIResponse "Unauthorized"
// @Failure 403 {object} api.APIResponse "Forbidden - not the organizer"
// @Failure 404 {object} api.APIResponse "Event not found"
// @Router /events/{id} [put]
func UpdateEvent(c *gin.Context, db *gorm.DB) {
	eventID := c.Param("id")

	var event models.Event
	if err := db.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, api.APIResponse{Error: "Event not found"})
		return
	}

	userID, _ := c.Get("user_id")
	if userID.(uint) != event.OrganizerID {
		c.JSON(http.StatusForbidden, api.APIResponse{Error: "You are not the organizer of this event"})
		return
	}

	var request api.UpdateEventRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Invalid payload"})
		return
	}

	if request.Name != nil {
		event.Name = *request.Name
	}
	if request.Description != nil {
		event.Description = *request.Description
	}
	if request.EventDateTime != nil {
		eventTime, err := time.Parse(time.RFC3339, *request.EventDateTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Invalid date format"})
			return
		}
		event.EventDateTime = eventTime
	}
	if request.Budget != nil {
		event.InitialBudget = *request.Budget
	}
	if request.Place != nil {
		event.Place = *request.Place
	}

	db.Save(&event)
	c.JSON(http.StatusOK, api.APIResponse{
		Message: "Event updated successfully",
		Data:    toEventResponse(&event),
	})
}

// @Summary Get event budget details
// @Description Get the budget details for an event, including initial budget, real budget, and difference
// @Tags events
// @Produce json
// @Security BearerAuth
// @Param id path int true "Event ID"
// @Success 200 {object} api.EventBudgetResponse "Budget details retrieved successfully"
// @Failure 401 {object} api.APIResponse "Unauthorized"
// @Failure 404 {object} api.APIResponse "Event not found"
// @Router /events/{id}/budget [get]
func GetEventBudget(c *gin.Context, db *gorm.DB) {
	var event models.Event
	id := c.Param("id")

	if err := db.Preload("Tasks").First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, api.APIResponse{Error: "Event not found"})
		return
	}

	realBudget := 0.0
	for _, task := range event.Tasks {
		if task.IsCompleted {
			realBudget += task.Budget
		}
	}

	c.JSON(http.StatusOK, api.EventBudgetResponse{
		InitialBudget: event.InitialBudget,
		RealBudget:    realBudget,
		Difference:    event.InitialBudget - realBudget,
	})
}

// @Summary Get event leaderboard
// @Description Get the leaderboard for an event
// @Tags events
// @Produce json
// @Security BearerAuth
// @Param id path int true "Event ID"
// @Success 200 {object} api.EventLeaderboardResponse "Leaderboard retrieved successfully"
// @Failure 401 {object} api.APIResponse "Unauthorized"
// @Failure 403 {object} api.APIResponse "Not a participant"
// @Failure 404 {object} api.APIResponse "Leaderboard not found"
// @Router /events/{id}/leaderboard [get]
func GetEventLeaderboard(c *gin.Context, db *gorm.DB) {
	userID, _ := c.Get("user_id")
	eventID := c.Param("id")

	// Check if the user is the organizer or a participant of the event
	var participation models.EventParticipation
	if err := db.Where("event_id = ? AND user_id = ?", eventID, userID).First(&participation).Error; err != nil {
		c.JSON(http.StatusForbidden, api.APIResponse{Error: "You are not a participant of this event"})
		return
	}

	var leaderboard []models.EventScore
	if err := db.Where("event_id = ?", eventID).Order("score DESC").Find(&leaderboard).Error; err != nil {
		c.JSON(http.StatusNotFound, api.APIResponse{Error: "Leaderboard not found"})
		return
	}

	var response api.EventLeaderboardResponse
	for _, entry := range leaderboard {
		response.Leaderboard = append(response.Leaderboard, api.EventLeaderboardEntry{
			UserID:  entry.UserID,
			Score:   entry.Score,
			EventID: entry.EventID,
		})
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Get user's events
// @Description Get all events where the user is a participant
// @Tags events
// @Produce json
// @Security BearerAuth
// @Success 200 {object} api.APIResponse{data=[]api.EventResponse} "Events retrieved successfully"
// @Failure 401 {object} api.APIResponse "Unauthorized"
// @Router /events [get]
func GetEvents(c *gin.Context, db *gorm.DB) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, api.APIResponse{Error: "Unauthorized"})
		return
	}

	var events []models.Event
	db.Joins("JOIN event_participations ON events.id = event_participations.event_id").
		Where("event_participations.user_id = ?", userID).
		Find(&events)

	var response []api.EventResponse
	for _, event := range events {
		if eventResponse := toEventResponse(&event); eventResponse != nil {
			response = append(response, *eventResponse)
		}
	}

	c.JSON(http.StatusOK, api.APIResponse{Data: response})
}

// @Summary Find best time slots for an event
// @Description Find the best available time slots for an event based on participants' schedules and specified time range
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body api.FindBestTimeSlotsRequest true "Find best time slots request details"
// @Success 200 {object} api.FindBestTimeSlotsResponse "Time slots found successfully"
// @Failure 400 {object} api.APIResponse "Invalid request"
// @Failure 403 {object} api.APIResponse "Forbidden - not a participant of the event"
// @Failure 404 {object} api.APIResponse "Event not found or no participants"
// @Router /events/find_best_time_for_day [post]
func FindBestTimeSlotsForDay(c *gin.Context, db *gorm.DB) {
	var request api.FindBestTimeSlotsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Invalid payload"})
		return
	}

	// Set default values for startTime and endTime if not provided
	if request.StartTime == "" {
		request.StartTime = leftTimeBoundForBusySlots
	}

	if request.EndTime == "" {
		request.EndTime = rightTimeBoundForBusySlots
	}

	// Validate time format
	_, startErr := time.Parse("15:04", request.StartTime)
	_, endErr := time.Parse("15:04", request.EndTime)

	if startErr != nil || endErr != nil {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Invalid time format. Use HH:MM format (24-hour)."})
		return
	}

	// Additional validation for time format
	if len(request.StartTime) != 5 || request.StartTime[2] != ':' {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Invalid start time format. Use HH:MM format (24-hour)."})
		return
	}

	if len(request.EndTime) != 5 || request.EndTime[2] != ':' {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Invalid end time format. Use HH:MM format (24-hour)."})
		return
	}

	var event models.Event
	if err := db.First(&event, request.EventID).Error; err != nil {
		c.JSON(http.StatusNotFound, api.APIResponse{Error: "Event not found"})
		return
	}

	// Check if the user is a participant or organizer of the event
	userID, _ := c.Get("user_id")
	if event.OrganizerID != userID.(uint) {
		var participation models.EventParticipation
		if err := db.Where("event_id = ? AND user_id = ?", request.EventID, userID).First(&participation).Error; err != nil {
			c.JSON(http.StatusForbidden, api.APIResponse{Error: "You are not a participant of this event"})
			return
		}
	}

	var participants []models.User
	db.Joins("JOIN event_participations ON event_participations.user_id = users.id").
		Where("event_participations.event_id = ?", request.EventID).
		Find(&participants)

	if len(participants) == 0 {
		c.JSON(http.StatusNotFound, api.APIResponse{Error: "No participants found"})
		return
	}

	busySlots := make(map[string]int)
	for _, user := range participants {
		getUserBusySlotsForDay(db, user.ID, request.Date, &busySlots)
	}

	suggestedSlots := suggestTimeSlotsForDay(&busySlots, request.Date, request.DurationMins, request.StartTime, request.EndTime)

	c.JSON(http.StatusOK, api.FindBestTimeSlotsResponse{
		Suggestions: suggestedSlots,
	})
}

// GetEventParticipants godoc
// @Summary Get event participants
// @Description Get a list of display names of all participants in an event
// @Tags events
// @Accept json
// @Produce json
// @Param id path int true "Event ID"
// @Security BearerAuth
// @Success 200 {object} api.EventParticipantsResponse "List of participants' display names"
// @Failure 400 {object} api.APIResponse "Invalid event ID"
// @Failure 401 {object} api.APIResponse "Unauthorized"
// @Failure 403 {object} api.APIResponse "Forbidden - not a participant of the event"
// @Failure 404 {object} api.APIResponse "Event not found"
// @Router /events/{id}/participants [get]
func GetEventParticipants(c *gin.Context, db *gorm.DB) {
	// Get the event ID from path parameter
	eventIDStr := c.Param("id")
	if eventIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Event ID is required"})
		return
	}

	// Parse event ID
	var eventID uint
	if _, err := fmt.Sscanf(eventIDStr, "%d", &eventID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID format"})
		return
	}

	// Get the current user ID from the context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Check if the event exists
	var event models.Event
	if err := db.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	// Check if the user is a participant of the event
	var participation models.EventParticipation
	if err := db.Where("event_id = ? AND user_id = ?", eventID, userID).First(&participation).Error; err != nil {
		// Also check if the user is the organizer
		if event.OrganizerID != userID.(uint) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not a participant of this event"})
			return
		}
	}

	// Get all participants of the event
	var participations []models.EventParticipation
	if err := db.Where("event_id = ?", eventID).Find(&participations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve event participants"})
		return
	}

	// Get the display names of all participants
	var participants []string

	// Add the organizer to the list
	var organizer models.User
	if err := db.First(&organizer, event.OrganizerID).Error; err == nil {
		participants = append(participants, organizer.DisplayName)
	}

	// Add other participants
	for _, p := range participations {
		// Skip if the participant is the organizer (already added)
		if p.UserID == event.OrganizerID {
			continue
		}

		var user models.User
		if err := db.First(&user, p.UserID).Error; err == nil {
			participants = append(participants, user.DisplayName)
		}
	}

	// Return the list of participants
	c.JSON(http.StatusOK, api.EventParticipantsResponse{
		Participants: participants,
	})
}

// GetEvent godoc
// @Summary Get event details
// @Description Get detailed information about a specific event
// @Tags events
// @Produce json
// @Security BearerAuth
// @Param id path int true "Event ID"
// @Success 200 {object} api.APIResponse{data=api.EventResponse} "Event details retrieved successfully"
// @Failure 401 {object} api.APIResponse "Unauthorized"
// @Failure 403 {object} api.APIResponse "Forbidden - not a participant of the event"
// @Failure 404 {object} api.APIResponse "Event not found"
// @Router /events/{id} [get]
func GetEvent(c *gin.Context, db *gorm.DB) {
	eventID := c.Param("id")

	var event models.Event
	if err := db.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, api.APIResponse{Error: "Event not found"})
		return
	}

	// Get the current user ID from the context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, api.APIResponse{Error: "User not authenticated"})
		return
	}

	// Check if the user is the organizer or a participant of the event
	if event.OrganizerID != userID.(uint) {
		var participation models.EventParticipation
		if err := db.Where("event_id = ? AND user_id = ?", event.ID, userID).First(&participation).Error; err != nil {
			c.JSON(http.StatusForbidden, api.APIResponse{Error: "You are not a participant of this event"})
			return
		}
	}

	c.JSON(http.StatusOK, api.APIResponse{
		Data: toEventResponse(&event),
	})
}
