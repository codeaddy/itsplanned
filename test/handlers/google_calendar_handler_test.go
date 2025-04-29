package handlers_test

import (
	"bytes"
	"encoding/json"
	"itsplanned/handlers"
	"itsplanned/test"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFetchGoogleCalendarEventsHandler(t *testing.T) {
	// Setup test database
	cleanup := test.SetupTestDB(t)
	defer cleanup()

	// Create test user
	user := test.CreateTestUser(t)

	testCases := []struct {
		name         string
		userID       uint
		payload      map[string]interface{}
		expectedCode int
		validateFunc func(t *testing.T, response map[string]interface{})
	}{
		{
			name:   "Fetch calendar events with valid dates",
			userID: user.ID,
			payload: map[string]interface{}{
				"access_token": "test_token",
				"start_date":   time.Now().Format(time.RFC3339),
				"end_date":     time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			},
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response map[string]interface{}) {
				// Should have events array
				events, ok := response["events"].([]interface{})
				assert.True(t, ok)
				assert.NotNil(t, events)
			},
		},
		{
			name:   "Fetch calendar events with invalid start date",
			userID: user.ID,
			payload: map[string]interface{}{
				"access_token": "test_token",
				"start_date":   "invalid_date",
				"end_date":     time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			},
			expectedCode: http.StatusBadRequest,
			validateFunc: nil,
		},
		{
			name:   "Fetch calendar events with invalid end date",
			userID: user.ID,
			payload: map[string]interface{}{
				"access_token": "test_token",
				"start_date":   time.Now().Format(time.RFC3339),
				"end_date":     "invalid_date",
			},
			expectedCode: http.StatusBadRequest,
			validateFunc: nil,
		},
		{
			name:   "Fetch calendar events with missing access token",
			userID: user.ID,
			payload: map[string]interface{}{
				"start_date": time.Now().Format(time.RFC3339),
				"end_date":   time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			},
			expectedCode: http.StatusBadRequest,
			validateFunc: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test context
			c, w := test.CreateTestContext(t, tc.userID)

			// Convert payload to JSON
			payloadJSON, err := json.Marshal(tc.payload)
			assert.NoError(t, err)

			// Create request
			c.Request = httptest.NewRequest("POST", "/calendar/events", bytes.NewBuffer(payloadJSON))
			c.Request.Header.Set("Content-Type", "application/json")

			// Call handler
			handlers.FetchGoogleCalendarEventsHandler(c)

			// Check status code
			assert.Equal(t, tc.expectedCode, w.Code)

			// If we expect a successful response, validate it
			if tc.expectedCode == http.StatusOK && tc.validateFunc != nil {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				tc.validateFunc(t, response)
			}
		})
	}
}
