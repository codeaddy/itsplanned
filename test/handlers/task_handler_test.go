package handlers_test

import (
	"encoding/json"
	"fmt"
	"itsplanned/handlers"
	"itsplanned/models/api"
	"itsplanned/test"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTasks(t *testing.T) {
	// Setup test database
	cleanup := test.SetupTestDB(t)
	defer cleanup()

	// Create test data
	user := test.CreateTestUser(t)
	event := test.CreateTestEvent(t, user.ID)
	_ = test.CreateTestTask(t, event.ID)
	_ = test.CreateTestTask(t, event.ID)

	// Test getting tasks as organizer
	t.Run("Organizer can get tasks", func(t *testing.T) {
		c, w := test.CreateTestContext(t, user.ID)
		c.Request = httptest.NewRequest("GET", fmt.Sprintf("/tasks?event_id=%d", event.ID), nil)
		c.Request.Header.Set("Content-Type", "application/json")

		handlers.GetTasks(c, test.TestDB)

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Check that we got 2 tasks back
		tasks, ok := response.Data.([]interface{})
		if assert.True(t, ok) {
			assert.Equal(t, 2, len(tasks))
		}
	})

	// Create another user who is not a participant
	otherUser := test.CreateTestUser(t)

	t.Run("Non-participant cannot get tasks", func(t *testing.T) {
		c, w := test.CreateTestContext(t, otherUser.ID)
		c.Request = httptest.NewRequest("GET", fmt.Sprintf("/tasks?event_id=%d", event.ID), nil)
		c.Request.Header.Set("Content-Type", "application/json")

		handlers.GetTasks(c, test.TestDB)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	// Add other user as participant and test again
	test.AddEventParticipant(t, event.ID, otherUser.ID)

	t.Run("Participant can get tasks", func(t *testing.T) {
		c, w := test.CreateTestContext(t, otherUser.ID)
		c.Request = httptest.NewRequest("GET", fmt.Sprintf("/tasks?event_id=%d", event.ID), nil)
		c.Request.Header.Set("Content-Type", "application/json")

		handlers.GetTasks(c, test.TestDB)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test with invalid event ID
	t.Run("Invalid event ID returns bad request", func(t *testing.T) {
		c, w := test.CreateTestContext(t, user.ID)
		c.Request = httptest.NewRequest("GET", "/tasks?event_id=invalid", nil)
		c.Request.Header.Set("Content-Type", "application/json")

		handlers.GetTasks(c, test.TestDB)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test with non-existent event ID
	t.Run("Non-existent event ID returns not found", func(t *testing.T) {
		c, w := test.CreateTestContext(t, user.ID)
		c.Request = httptest.NewRequest("GET", "/tasks?event_id=999", nil)
		c.Request.Header.Set("Content-Type", "application/json")

		handlers.GetTasks(c, test.TestDB)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
