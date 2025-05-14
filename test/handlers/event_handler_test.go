package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"itsplanned/handlers"
	"itsplanned/models"
	"itsplanned/models/api"
	"itsplanned/test"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestFindBestTimeSlotsForDay(t *testing.T) {
	cleanup := test.SetupTestDB(t)
	defer cleanup()

	organizer := test.CreateTestUser(t)
	participant1 := test.CreateTestUser(t)
	participant2 := test.CreateTestUser(t)

	event := test.CreateTestEvent(t, organizer.ID)

	test.AddEventParticipant(t, event.ID, participant1.ID)
	test.AddEventParticipant(t, event.ID, participant2.ID)

	today := time.Now().Format("2006-01-02")

	createCalendarEvent(t, participant1.ID, today, "10:00", "11:30", "Meeting")

	createCalendarEvent(t, participant2.ID, today, "12:00", "13:00", "Lunch")

	testCases := []struct {
		name         string
		userID       uint
		request      api.FindBestTimeSlotsRequest
		expectedCode int
		validateFunc func(t *testing.T, response *api.FindBestTimeSlotsResponse)
	}{
		{
			name:   "Valid request with default time range",
			userID: organizer.ID,
			request: api.FindBestTimeSlotsRequest{
				EventID:      event.ID,
				Date:         today,
				DurationMins: 60,
			},
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response *api.FindBestTimeSlotsResponse) {
				assert.Greater(t, len(response.Suggestions), 0)

				assert.LessOrEqual(t, response.Suggestions[0].BusyCount, response.Suggestions[len(response.Suggestions)-1].BusyCount)
			},
		},
		{
			name:   "Valid request with specific time range",
			userID: organizer.ID,
			request: api.FindBestTimeSlotsRequest{
				EventID:      event.ID,
				Date:         today,
				DurationMins: 60,
				StartTime:    "09:00",
				EndTime:      "14:00",
			},
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response *api.FindBestTimeSlotsResponse) {
				assert.Greater(t, len(response.Suggestions), 0)

				firstSlotTime := response.Suggestions[0].Slot[len(today)+1:]
				assert.GreaterOrEqual(t, firstSlotTime, "09:00")

				if len(response.Suggestions) > 0 {
					lastSlot := response.Suggestions[len(response.Suggestions)-1].Slot
					lastSlotTime := lastSlot[len(today)+1:]
					lastSlotTimeParsed, _ := time.Parse("15:04", lastSlotTime)
					endTimeParsed, _ := time.Parse("15:04", "14:00")

					assert.True(t, lastSlotTimeParsed.Before(endTimeParsed) || lastSlotTimeParsed.Equal(endTimeParsed.Add(-time.Hour)))
				}
			},
		},
		{
			name:   "Non-participant tries to find time slots",
			userID: test.CreateTestUser(t).ID,
			request: api.FindBestTimeSlotsRequest{
				EventID:      event.ID,
				Date:         today,
				DurationMins: 60,
			},
			expectedCode: http.StatusForbidden,
			validateFunc: nil,
		},
		{
			name:   "Invalid time format",
			userID: organizer.ID,
			request: api.FindBestTimeSlotsRequest{
				EventID:      event.ID,
				Date:         today,
				DurationMins: 60,
				StartTime:    "9:00", // Missing leading zero
				EndTime:      "14:00",
			},
			expectedCode: http.StatusBadRequest,
			validateFunc: nil,
		},
		{
			name:   "Non-existent event",
			userID: organizer.ID,
			request: api.FindBestTimeSlotsRequest{
				EventID:      9999,
				Date:         today,
				DurationMins: 60,
			},
			expectedCode: http.StatusNotFound,
			validateFunc: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := test.CreateTestContext(t, tc.userID)

			requestJSON, err := json.Marshal(tc.request)
			assert.NoError(t, err)

			c.Request = httptest.NewRequest("POST", "/events/find_best_time_for_day", bytes.NewBuffer(requestJSON))
			c.Request.Header.Set("Content-Type", "application/json")

			handlers.FindBestTimeSlotsForDay(c, test.TestDB)

			assert.Equal(t, tc.expectedCode, w.Code)

			if tc.expectedCode == http.StatusOK && tc.validateFunc != nil {
				var response api.FindBestTimeSlotsResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				tc.validateFunc(t, &response)
			}
		})
	}
}

func createCalendarEvent(t *testing.T, userID uint, date, startTime, endTime, summary string) {
	startDateTime, err := time.Parse("2006-01-02 15:04", fmt.Sprintf("%s %s", date, startTime))
	assert.NoError(t, err)

	endDateTime, err := time.Parse("2006-01-02 15:04", fmt.Sprintf("%s %s", date, endTime))
	assert.NoError(t, err)

	calendarEvent := &models.CalendarEvent{
		UserID:    userID,
		Title:     summary,
		StartTime: startDateTime,
		EndTime:   endDateTime,
	}

	err = test.TestDB.Create(calendarEvent).Error
	assert.NoError(t, err)
}

func TestCreateEvent(t *testing.T) {
	cleanup := test.SetupTestDB(t)
	defer cleanup()

	user := test.CreateTestUser(t)

	testCases := []struct {
		name         string
		userID       uint
		request      api.CreateEventRequest
		expectedCode int
		validateFunc func(t *testing.T, response api.APIResponse)
	}{
		{
			name:   "Successfully create event",
			userID: user.ID,
			request: api.CreateEventRequest{
				Name:          "Test Birthday Party",
				Description:   "Celebrating a birthday",
				EventDateTime: time.Now().Add(24 * time.Hour).Format(time.RFC3339),
				InitialBudget: 1000.0,
				Place:         "Test Location",
			},
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response api.APIResponse) {
				assert.Equal(t, "Event created", response.Message)

				eventData, ok := response.Data.(map[string]interface{})
				assert.True(t, ok)

				assert.Equal(t, "Test Birthday Party", eventData["name"])
				assert.Equal(t, "Celebrating a birthday", eventData["description"])
				assert.Equal(t, float64(1000.0), eventData["initial_budget"])
				assert.Equal(t, "Test Location", eventData["place"])
				assert.Equal(t, float64(user.ID), eventData["organizer_id"])

				var count int64
				test.TestDB.Model(&models.Event{}).Where("name = ? AND organizer_id = ?", "Test Birthday Party", user.ID).Count(&count)
				assert.Equal(t, int64(1), count)

				var participationCount int64
				var event models.Event
				test.TestDB.Where("name = ? AND organizer_id = ?", "Test Birthday Party", user.ID).First(&event)
				test.TestDB.Model(&models.EventParticipation{}).Where("event_id = ? AND user_id = ?", event.ID, user.ID).Count(&participationCount)
				assert.Equal(t, int64(1), participationCount)
			},
		},
		{
			name:   "Invalid request - missing required fields",
			userID: user.ID,
			request: api.CreateEventRequest{
				Description:   "Celebrating a birthday",
				InitialBudget: 1000.0,
				Place:         "Test Location",
			},
			expectedCode: http.StatusBadRequest,
			validateFunc: func(t *testing.T, response api.APIResponse) {
				assert.Contains(t, response.Error, "Invalid payload")
			},
		},
		{
			name:   "Invalid date format",
			userID: user.ID,
			request: api.CreateEventRequest{
				Name:          "Test Birthday Party",
				Description:   "Celebrating a birthday",
				EventDateTime: "2023-01-01", // Invalid format, should be RFC3339
				InitialBudget: 1000.0,
				Place:         "Test Location",
			},
			expectedCode: http.StatusBadRequest,
			validateFunc: func(t *testing.T, response api.APIResponse) {
				assert.Contains(t, response.Error, "Invalid date format")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := test.CreateTestContext(t, tc.userID)

			requestJSON, err := json.Marshal(tc.request)
			assert.NoError(t, err)

			c.Request = httptest.NewRequest("POST", "/events", bytes.NewBuffer(requestJSON))
			c.Request.Header.Set("Content-Type", "application/json")

			handlers.CreateEvent(c, test.TestDB)

			assert.Equal(t, tc.expectedCode, w.Code)

			var response api.APIResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tc.validateFunc != nil {
				tc.validateFunc(t, response)
			}
		})
	}
}

func TestGetEvent(t *testing.T) {
	cleanup := test.SetupTestDB(t)
	defer cleanup()

	organizer := test.CreateTestUser(t)
	event := test.CreateTestEvent(t, organizer.ID)

	participant := test.CreateTestUser(t)
	test.AddEventParticipant(t, event.ID, participant.ID)

	nonParticipant := test.CreateTestUser(t)

	testCases := []struct {
		name         string
		userID       uint
		eventID      uint
		expectedCode int
		validateFunc func(t *testing.T, response api.APIResponse)
	}{
		{
			name:         "Organizer can view event",
			userID:       organizer.ID,
			eventID:      event.ID,
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response api.APIResponse) {
				eventData, ok := response.Data.(map[string]interface{})
				assert.True(t, ok)

				assert.Equal(t, float64(event.ID), eventData["id"])
				assert.Equal(t, event.Name, eventData["name"])
				assert.Equal(t, event.Description, eventData["description"])
				assert.Equal(t, event.InitialBudget, eventData["initial_budget"])
				assert.Equal(t, event.Place, eventData["place"])
				assert.Equal(t, float64(organizer.ID), eventData["organizer_id"])
			},
		},
		{
			name:         "Participant can view event",
			userID:       participant.ID,
			eventID:      event.ID,
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response api.APIResponse) {
				eventData, ok := response.Data.(map[string]interface{})
				assert.True(t, ok)

				assert.Equal(t, float64(event.ID), eventData["id"])
				assert.Equal(t, event.Name, eventData["name"])
			},
		},
		{
			name:         "Non-participant cannot view event",
			userID:       nonParticipant.ID,
			eventID:      event.ID,
			expectedCode: http.StatusForbidden,
			validateFunc: func(t *testing.T, response api.APIResponse) {
				assert.Contains(t, response.Error, "You are not a participant of this event")
			},
		},
		{
			name:         "Event not found",
			userID:       organizer.ID,
			eventID:      9999,
			expectedCode: http.StatusNotFound,
			validateFunc: func(t *testing.T, response api.APIResponse) {
				assert.Contains(t, response.Error, "Event not found")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := test.CreateTestContext(t, tc.userID)

			c.Request = httptest.NewRequest("GET", fmt.Sprintf("/events/%d", tc.eventID), nil)
			c.Params = []gin.Param{{Key: "id", Value: fmt.Sprintf("%d", tc.eventID)}}

			handlers.GetEvent(c, test.TestDB)

			assert.Equal(t, tc.expectedCode, w.Code)

			var response api.APIResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tc.validateFunc != nil {
				tc.validateFunc(t, response)
			}
		})
	}
}

func TestUpdateEvent(t *testing.T) {
	cleanup := test.SetupTestDB(t)
	defer cleanup()

	organizer := test.CreateTestUser(t)
	event := test.CreateTestEvent(t, organizer.ID)

	participant := test.CreateTestUser(t)
	test.AddEventParticipant(t, event.ID, participant.ID)

	strPtr := func(s string) *string { return &s }
	floatPtr := func(f float64) *float64 { return &f }

	testCases := []struct {
		name         string
		userID       uint
		eventID      uint
		request      api.UpdateEventRequest
		expectedCode int
		validateFunc func(t *testing.T, response api.APIResponse)
	}{
		{
			name:    "Successfully update event as organizer",
			userID:  organizer.ID,
			eventID: event.ID,
			request: api.UpdateEventRequest{
				Name:          strPtr("Updated Event Name"),
				Description:   strPtr("Updated Description"),
				EventDateTime: strPtr(time.Now().Add(48 * time.Hour).Format(time.RFC3339)),
				Budget:        floatPtr(2000.0),
				Place:         strPtr("Updated Location"),
			},
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response api.APIResponse) {
				eventData, ok := response.Data.(map[string]interface{})
				assert.True(t, ok)

				assert.Equal(t, "Updated Event Name", eventData["name"])
				assert.Equal(t, "Updated Description", eventData["description"])
				assert.Equal(t, float64(2000.0), eventData["initial_budget"])
				assert.Equal(t, "Updated Location", eventData["place"])

				var updatedEvent models.Event
				err := test.TestDB.First(&updatedEvent, event.ID).Error
				assert.NoError(t, err)

				assert.Equal(t, "Updated Event Name", updatedEvent.Name)
				assert.Equal(t, "Updated Description", updatedEvent.Description)
				assert.Equal(t, float64(2000.0), updatedEvent.InitialBudget)
				assert.Equal(t, "Updated Location", updatedEvent.Place)
			},
		},
		{
			name:    "Partial update with only some fields",
			userID:  organizer.ID,
			eventID: event.ID,
			request: api.UpdateEventRequest{
				Name:  strPtr("Only Name Updated"),
				Place: strPtr("Only Place Updated"),
			},
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response api.APIResponse) {
				var updatedEvent models.Event
				err := test.TestDB.First(&updatedEvent, event.ID).Error
				assert.NoError(t, err)

				assert.Equal(t, "Only Name Updated", updatedEvent.Name)
				assert.Equal(t, "Only Place Updated", updatedEvent.Place)
				assert.Equal(t, "Updated Description", updatedEvent.Description)
				assert.Equal(t, float64(2000.0), updatedEvent.InitialBudget)
			},
		},
		{
			name:    "Non-organizer cannot update event",
			userID:  participant.ID,
			eventID: event.ID,
			request: api.UpdateEventRequest{
				Name: strPtr("Should Not Update"),
			},
			expectedCode: http.StatusForbidden,
			validateFunc: func(t *testing.T, response api.APIResponse) {
				assert.Contains(t, response.Error, "You are not the organizer of this event")
			},
		},
		{
			name:    "Event not found",
			userID:  organizer.ID,
			eventID: 9999,
			request: api.UpdateEventRequest{
				Name: strPtr("Should Not Update"),
			},
			expectedCode: http.StatusNotFound,
			validateFunc: func(t *testing.T, response api.APIResponse) {
				assert.Contains(t, response.Error, "Event not found")
			},
		},
		{
			name:    "Invalid date format",
			userID:  organizer.ID,
			eventID: event.ID,
			request: api.UpdateEventRequest{
				EventDateTime: strPtr("2023-01-01"), // Invalid format, should be RFC3339
			},
			expectedCode: http.StatusBadRequest,
			validateFunc: func(t *testing.T, response api.APIResponse) {
				assert.Contains(t, response.Error, "Invalid date format")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := test.CreateTestContext(t, tc.userID)

			requestJSON, err := json.Marshal(tc.request)
			assert.NoError(t, err)

			c.Request = httptest.NewRequest("PUT", fmt.Sprintf("/events/%d", tc.eventID), bytes.NewBuffer(requestJSON))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = []gin.Param{{Key: "id", Value: fmt.Sprintf("%d", tc.eventID)}}

			handlers.UpdateEvent(c, test.TestDB)

			assert.Equal(t, tc.expectedCode, w.Code)

			var response api.APIResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tc.validateFunc != nil {
				tc.validateFunc(t, response)
			}
		})
	}
}

func TestDeleteEvent(t *testing.T) {
	cleanup := test.SetupTestDB(t)
	defer cleanup()

	organizer := test.CreateTestUser(t)
	event := test.CreateTestEvent(t, organizer.ID)

	participant := test.CreateTestUser(t)
	test.AddEventParticipant(t, event.ID, participant.ID)

	task1 := test.CreateTestTask(t, event.ID)
	task2 := test.CreateTestTask(t, event.ID)

	task1.AssignedTo = &participant.ID
	test.TestDB.Save(&task1)

	taskEvent := &models.TaskStatusEvent{
		TaskID:        task1.ID,
		TaskName:      task1.Title,
		OldStatus:     "unassigned",
		NewStatus:     "assigned",
		UserID:        organizer.ID,
		ChangedByID:   participant.ID,
		ChangedByName: "Test User",
		IsRead:        false,
		EventTime:     time.Now(),
	}
	test.TestDB.Create(&taskEvent)

	eventScore := &models.EventScore{
		EventID: event.ID,
		UserID:  participant.ID,
		Score:   10.0,
	}
	test.TestDB.Create(&eventScore)

	invite := &models.EventInvitation{
		EventID:    event.ID,
		InviteCode: "test_invite_code",
	}
	test.TestDB.Create(&invite)

	testCases := []struct {
		name         string
		userID       uint
		eventID      uint
		expectedCode int
		validateFunc func(t *testing.T, response api.APIResponse)
	}{
		{
			name:         "Non-organizer cannot delete event",
			userID:       participant.ID,
			eventID:      event.ID,
			expectedCode: http.StatusForbidden,
			validateFunc: func(t *testing.T, response api.APIResponse) {
				assert.Contains(t, response.Error, "You are not the organizer of this event")

				var count int64
				test.TestDB.Model(&models.Event{}).Where("id = ?", event.ID).Count(&count)
				assert.Equal(t, int64(1), count)
			},
		},
		{
			name:         "Event not found",
			userID:       organizer.ID,
			eventID:      9999,
			expectedCode: http.StatusNotFound,
			validateFunc: func(t *testing.T, response api.APIResponse) {
				assert.Contains(t, response.Error, "Event not found")
			},
		},
		{
			name:         "Successfully delete event and all associated data",
			userID:       organizer.ID,
			eventID:      event.ID,
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response api.APIResponse) {
				assert.Contains(t, response.Message, "Event and all associated data deleted successfully")

				var eventCount int64
				test.TestDB.Model(&models.Event{}).Where("id = ?", event.ID).Count(&eventCount)
				assert.Equal(t, int64(0), eventCount)

				var taskCount int64
				test.TestDB.Model(&models.Task{}).Where("event_id = ?", event.ID).Count(&taskCount)
				assert.Equal(t, int64(0), taskCount)

				var taskEventCount int64
				test.TestDB.Model(&models.TaskStatusEvent{}).Where("task_id IN (?)", []uint{task1.ID, task2.ID}).Count(&taskEventCount)
				assert.Equal(t, int64(0), taskEventCount)

				var scoreCount int64
				test.TestDB.Model(&models.EventScore{}).Where("event_id = ?", event.ID).Count(&scoreCount)
				assert.Equal(t, int64(0), scoreCount)

				var participationCount int64
				test.TestDB.Model(&models.EventParticipation{}).Where("event_id = ?", event.ID).Count(&participationCount)
				assert.Equal(t, int64(0), participationCount)

				var inviteCount int64
				test.TestDB.Model(&models.EventInvitation{}).Where("event_id = ?", event.ID).Count(&inviteCount)
				assert.Equal(t, int64(0), inviteCount)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := test.CreateTestContext(t, tc.userID)

			c.Request = httptest.NewRequest("DELETE", fmt.Sprintf("/events/%d", tc.eventID), nil)
			c.Params = []gin.Param{{Key: "id", Value: fmt.Sprintf("%d", tc.eventID)}}

			handlers.DeleteEvent(c, test.TestDB)

			assert.Equal(t, tc.expectedCode, w.Code)

			var response api.APIResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tc.validateFunc != nil {
				tc.validateFunc(t, response)
			}
		})
	}
}
