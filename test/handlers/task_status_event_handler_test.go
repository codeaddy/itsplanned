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
	cleanup := test.SetupTestDB(t)
	defer cleanup()

	user := test.CreateTestUser(t)
	otherUser := test.CreateTestUser(t)

	event := test.CreateTestEvent(t, user.ID)
	task := test.CreateTestTask(t, event.ID)

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
				assert.Equal(t, 2, len(response.Events))

				assert.True(t, response.Events[0].EventTime.After(response.Events[1].EventTime))

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
				assert.Equal(t, 1, len(response.Events))

				assert.Equal(t, otherUser.ID, response.Events[0].ChangedByID)
			},
		},
		{
			name:         "Get unread events for non-existent user",
			userID:       9999,
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response *api.TaskStatusEventsResponse) {
				assert.Equal(t, 0, len(response.Events))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := test.CreateTestContext(t, tc.userID)

			handlers.GetUnreadTaskStatusEvents(c, test.TestDB)

			assert.Equal(t, tc.expectedCode, w.Code)

			if tc.expectedCode == http.StatusOK && tc.validateFunc != nil {
				var response api.TaskStatusEventsResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				tc.validateFunc(t, &response)

				var unreadCount int64
				test.TestDB.Model(&models.TaskStatusEvent{}).
					Where("user_id = ? AND is_read = ?", tc.userID, false).
					Count(&unreadCount)
				assert.Equal(t, int64(0), unreadCount)
			}
		})
	}
}

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
