package handlers_test

import (
	"encoding/json"
	"itsplanned/handlers"
	"itsplanned/models"
	"itsplanned/models/api"
	"itsplanned/test"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetUnreadTaskStatusEvents(t *testing.T) {
	// Setup test database
	cleanup := test.SetupTestDB(t)
	defer cleanup()

	// Create test users
	user := test.CreateTestUser(t)
	otherUser := test.CreateTestUser(t)

	// Create test event and task
	event := test.CreateTestEvent(t, user.ID)
	task := test.CreateTestTask(t, event.ID)

	// Create test task status events
	createTaskStatusEvent(t, user.ID, task.ID, "unassigned", "assigned", user.ID)
	createTaskStatusEvent(t, user.ID, task.ID, "assigned", "completed", user.ID)
	createTaskStatusEvent(t, otherUser.ID, task.ID, "unassigned", "assigned", otherUser.ID)

	testCases := []struct {
		name         string
		userID       uint
		expectedCode int
		validateFunc func(t *testing.T, response *api.TaskStatusEventsResponse)
	}{
		{
			name:         "Get unread events for user",
			userID:       user.ID,
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response *api.TaskStatusEventsResponse) {
				// Should have 2 unread events
				assert.Equal(t, 2, len(response.Events))

				// Events should be ordered by event_time DESC
				assert.True(t, response.Events[0].EventTime.After(response.Events[1].EventTime))

				// All events should be for the correct user
				for _, event := range response.Events {
					assert.Equal(t, user.ID, event.ChangedByID)
				}
			},
		},
		{
			name:         "Get unread events for other user",
			userID:       otherUser.ID,
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response *api.TaskStatusEventsResponse) {
				// Should have 1 unread event
				assert.Equal(t, 1, len(response.Events))

				// Event should be for the other user
				assert.Equal(t, otherUser.ID, response.Events[0].ChangedByID)
			},
		},
		{
			name:         "Get unread events for non-existent user",
			userID:       9999,
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response *api.TaskStatusEventsResponse) {
				// Should have no events
				assert.Equal(t, 0, len(response.Events))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test context
			c, w := test.CreateTestContext(t, tc.userID)

			// Call handler
			handlers.GetUnreadTaskStatusEvents(c, test.TestDB)

			// Check status code
			assert.Equal(t, tc.expectedCode, w.Code)

			// If we expect a successful response, validate it
			if tc.expectedCode == http.StatusOK && tc.validateFunc != nil {
				var response api.TaskStatusEventsResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				tc.validateFunc(t, &response)

				// Verify that events were marked as read
				var unreadCount int64
				test.TestDB.Model(&models.TaskStatusEvent{}).
					Where("user_id = ? AND is_read = ?", tc.userID, false).
					Count(&unreadCount)
				assert.Equal(t, int64(0), unreadCount)
			}
		})
	}
}

// Helper function to create a task status event for testing
func createTaskStatusEvent(t *testing.T, userID, taskID uint, oldStatus, newStatus string, changedByID uint) {
	event := &models.TaskStatusEvent{
		UserID:        userID,
		TaskID:        taskID,
		TaskName:      "Test Task",
		OldStatus:     oldStatus,
		NewStatus:     newStatus,
		ChangedByID:   changedByID,
		ChangedByName: "Test User",
		IsRead:        false,
		EventTime:     time.Now(),
	}

	err := test.TestDB.Create(event).Error
	assert.NoError(t, err)
}
