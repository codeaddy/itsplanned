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

	"github.com/stretchr/testify/assert"
)

func TestFindBestTimeSlotsForDay(t *testing.T) {
	// Setup test database
	cleanup := test.SetupTestDB(t)
	defer cleanup()

	// Create test users
	organizer := test.CreateTestUser(t)
	participant1 := test.CreateTestUser(t)
	participant2 := test.CreateTestUser(t)

	// Create test event
	event := test.CreateTestEvent(t, organizer.ID)

	// Add participants to the event
	test.AddEventParticipant(t, event.ID, participant1.ID)
	test.AddEventParticipant(t, event.ID, participant2.ID)

	// Create calendar events for participants
	today := time.Now().Format("2006-01-02")

	// Participant 1 has a meeting from 10:00 to 11:30
	createCalendarEvent(t, participant1.ID, today, "10:00", "11:30", "Meeting")

	// Participant 2 has a lunch from 12:00 to 13:00
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
				// We should get suggestions
				assert.Greater(t, len(response.Suggestions), 0)

				// The first suggestion should have the lowest busy count
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
				// We should get suggestions
				assert.Greater(t, len(response.Suggestions), 0)

				// First slot should be 09:00 or later
				firstSlotTime := response.Suggestions[0].Slot[len(today)+1:]
				assert.GreaterOrEqual(t, firstSlotTime, "09:00")

				// Last suggestion's slot should be before 14:00
				if len(response.Suggestions) > 0 {
					lastSlot := response.Suggestions[len(response.Suggestions)-1].Slot
					lastSlotTime := lastSlot[len(today)+1:]
					lastSlotTimeParsed, _ := time.Parse("15:04", lastSlotTime)
					endTimeParsed, _ := time.Parse("15:04", "14:00")

					// 60 minutes duration means the last valid start time is 13:00
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
			// Create test context
			c, w := test.CreateTestContext(t, tc.userID)

			// Convert request to JSON
			requestJSON, err := json.Marshal(tc.request)
			assert.NoError(t, err)

			// Create request
			c.Request = httptest.NewRequest("POST", "/events/find_best_time_for_day", bytes.NewBuffer(requestJSON))
			c.Request.Header.Set("Content-Type", "application/json")

			// Call handler
			handlers.FindBestTimeSlotsForDay(c, test.TestDB)

			// Check status code
			assert.Equal(t, tc.expectedCode, w.Code)

			// If we expect a successful response, validate it
			if tc.expectedCode == http.StatusOK && tc.validateFunc != nil {
				var response api.FindBestTimeSlotsResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				tc.validateFunc(t, &response)
			}
		})
	}
}

// Helper function to create a calendar event for testing
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
