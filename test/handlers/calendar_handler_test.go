package handlers_test

import (
	"encoding/json"
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

func TestImportCalendarEvents(t *testing.T) {
	// Setup test database
	cleanup := test.SetupTestDB(t)
	defer cleanup()

	// Create test user
	user := test.CreateTestUser(t)

	// Create test token
	token := &models.UserToken{
		UserID:       user.ID,
		AccessToken:  "test_encrypted_token",
		RefreshToken: "test_refresh_token",
		Expiry:       time.Now().Add(24 * time.Hour).Format(time.RFC3339),
	}
	err := test.TestDB.Create(token).Error
	assert.NoError(t, err)

	testCases := []struct {
		name         string
		userID       uint
		expectedCode int
		validateFunc func(t *testing.T, response *api.ImportCalendarEventsResponse)
	}{
		{
			name:         "Import calendar events with valid token",
			userID:       user.ID,
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response *api.ImportCalendarEventsResponse) {
				// Should have success message
				assert.Equal(t, "Events imported successfully", response.Message)

				// Verify events were imported
				var events []models.CalendarEvent
				err := test.TestDB.Where("user_id = ?", user.ID).Find(&events).Error
				assert.NoError(t, err)
				assert.Greater(t, len(events), 0)

				// Verify event times are in MSK timezone
				for _, event := range events {
					assert.Equal(t, "MSK", event.StartTime.Location().String())
					assert.Equal(t, "MSK", event.EndTime.Location().String())
				}
			},
		},
		{
			name:         "Import calendar events without token",
			userID:       test.CreateTestUser(t).ID, // New user without token
			expectedCode: http.StatusUnauthorized,
			validateFunc: nil,
		},
		{
			name:         "Import calendar events with invalid user ID",
			userID:       9999,
			expectedCode: http.StatusUnauthorized,
			validateFunc: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test context
			c, w := test.CreateTestContext(t, tc.userID)

			// Create request
			c.Request = httptest.NewRequest("GET", "/calendar/import", nil)

			// Call handler
			handlers.ImportCalendarEvents(c, test.TestDB)

			// Check status code
			assert.Equal(t, tc.expectedCode, w.Code)

			// If we expect a successful response, validate it
			if tc.expectedCode == http.StatusOK && tc.validateFunc != nil {
				var response api.ImportCalendarEventsResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				tc.validateFunc(t, &response)
			}
		})
	}
}
