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

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetTasks(t *testing.T) {
	cleanup := test.SetupTestDB(t)
	defer cleanup()

	user := test.CreateTestUser(t)
	event := test.CreateTestEvent(t, user.ID)
	_ = test.CreateTestTask(t, event.ID)
	_ = test.CreateTestTask(t, event.ID)

	t.Run("Organizer can get tasks", func(t *testing.T) {
		c, w := test.CreateTestContext(t, user.ID)
		c.Request = httptest.NewRequest("GET", fmt.Sprintf("/tasks?event_id=%d", event.ID), nil)
		c.Request.Header.Set("Content-Type", "application/json")

		handlers.GetTasks(c, test.TestDB)

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		tasks, ok := response.Data.([]interface{})
		if assert.True(t, ok) {
			assert.Equal(t, 2, len(tasks))
		}
	})

	otherUser := test.CreateTestUser(t)

	t.Run("Non-participant cannot get tasks", func(t *testing.T) {
		c, w := test.CreateTestContext(t, otherUser.ID)
		c.Request = httptest.NewRequest("GET", fmt.Sprintf("/tasks?event_id=%d", event.ID), nil)
		c.Request.Header.Set("Content-Type", "application/json")

		handlers.GetTasks(c, test.TestDB)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	test.AddEventParticipant(t, event.ID, otherUser.ID)

	t.Run("Participant can get tasks", func(t *testing.T) {
		c, w := test.CreateTestContext(t, otherUser.ID)
		c.Request = httptest.NewRequest("GET", fmt.Sprintf("/tasks?event_id=%d", event.ID), nil)
		c.Request.Header.Set("Content-Type", "application/json")

		handlers.GetTasks(c, test.TestDB)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Invalid event ID returns bad request", func(t *testing.T) {
		c, w := test.CreateTestContext(t, user.ID)
		c.Request = httptest.NewRequest("GET", "/tasks?event_id=invalid", nil)
		c.Request.Header.Set("Content-Type", "application/json")

		handlers.GetTasks(c, test.TestDB)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Non-existent event ID returns not found", func(t *testing.T) {
		c, w := test.CreateTestContext(t, user.ID)
		c.Request = httptest.NewRequest("GET", "/tasks?event_id=999", nil)
		c.Request.Header.Set("Content-Type", "application/json")

		handlers.GetTasks(c, test.TestDB)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestCreateTask(t *testing.T) {
	cleanup := test.SetupTestDB(t)
	defer cleanup()

	user := test.CreateTestUser(t)
	event := test.CreateTestEvent(t, user.ID)

	testCases := []struct {
		name         string
		userID       uint
		request      api.CreateTaskRequest
		expectedCode int
		validateFunc func(t *testing.T, response api.APIResponse)
	}{
		{
			name:   "Successfully create task",
			userID: user.ID,
			request: api.CreateTaskRequest{
				Title:       "New Test Task",
				Description: "Test Description",
				Budget:      200.0,
				Points:      20,
				EventID:     event.ID,
			},
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response api.APIResponse) {
				assert.Equal(t, "Task created", response.Message)

				taskData, ok := response.Data.(map[string]interface{})
				assert.True(t, ok)

				assert.Equal(t, "New Test Task", taskData["title"])
				assert.Equal(t, "Test Description", taskData["description"])
				assert.Equal(t, float64(200.0), taskData["budget"])
				assert.Equal(t, float64(20), taskData["points"])
				assert.Equal(t, float64(event.ID), taskData["event_id"])
				assert.Equal(t, false, taskData["is_completed"])

				var count int64
				test.TestDB.Model(&models.Task{}).Where("title = ? AND event_id = ?", "New Test Task", event.ID).Count(&count)
				assert.Equal(t, int64(1), count)
			},
		},
		{
			name:   "Non-organizer can create task",
			userID: test.CreateTestUser(t).ID,
			request: api.CreateTaskRequest{
				Title:       "Another Test Task",
				Description: "Another Description",
				Budget:      150.0,
				Points:      15,
				EventID:     event.ID,
			},
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response api.APIResponse) {
				assert.Equal(t, "Task created", response.Message)

				var count int64
				test.TestDB.Model(&models.Task{}).Where("title = ? AND event_id = ?", "Another Test Task", event.ID).Count(&count)
				assert.Equal(t, int64(1), count)
			},
		},
		{
			name:   "Invalid request - missing required fields",
			userID: user.ID,
			request: api.CreateTaskRequest{
				Description: "Test Description",
				Budget:      100.0,
				EventID:     event.ID,
			},
			expectedCode: http.StatusBadRequest,
			validateFunc: func(t *testing.T, response api.APIResponse) {
				assert.Contains(t, response.Error, "Invalid payload")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := test.CreateTestContext(t, tc.userID)

			requestJSON, err := json.Marshal(tc.request)
			assert.NoError(t, err)

			c.Request = httptest.NewRequest("POST", "/tasks", bytes.NewBuffer(requestJSON))
			c.Request.Header.Set("Content-Type", "application/json")

			handlers.CreateTask(c, test.TestDB)

			assert.Equal(t, tc.expectedCode, w.Code)

			if tc.expectedCode == http.StatusOK && tc.validateFunc != nil {
				var response api.APIResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				tc.validateFunc(t, response)
			}
		})
	}
}

func TestUpdateTask(t *testing.T) {
	cleanup := test.SetupTestDB(t)
	defer cleanup()

	user := test.CreateTestUser(t)
	event := test.CreateTestEvent(t, user.ID)
	task := test.CreateTestTask(t, event.ID)

	otherUser := test.CreateTestUser(t)
	test.AddEventParticipant(t, event.ID, otherUser.ID)

	strPtr := func(s string) *string { return &s }
	floatPtr := func(f float64) *float64 { return &f }
	intPtr := func(i int) *int { return &i }

	testCases := []struct {
		name         string
		userID       uint
		taskID       uint
		request      api.UpdateTaskRequest
		expectedCode int
		validateFunc func(t *testing.T, response api.APIResponse)
	}{
		{
			name:   "Successfully update task as organizer",
			userID: user.ID,
			taskID: task.ID,
			request: api.UpdateTaskRequest{
				Title:       strPtr("Updated Task Title"),
				Description: strPtr("Updated Description"),
				Budget:      floatPtr(300.0),
				Points:      intPtr(25),
			},
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response api.APIResponse) {
				assert.Equal(t, "Task updated successfully", response.Message)

				var updatedTask models.Task
				err := test.TestDB.First(&updatedTask, task.ID).Error
				assert.NoError(t, err)

				assert.Equal(t, "Updated Task Title", updatedTask.Title)
				assert.Equal(t, "Updated Description", updatedTask.Description)
				assert.Equal(t, 300.0, updatedTask.Budget)
				assert.Equal(t, 25, updatedTask.Points)
			},
		},
		{
			name:   "Partial update with only some fields",
			userID: user.ID,
			taskID: task.ID,
			request: api.UpdateTaskRequest{
				Title:  strPtr("Only Title Updated"),
				Points: intPtr(30),
			},
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response api.APIResponse) {
				assert.Equal(t, "Task updated successfully", response.Message)

				var updatedTask models.Task
				err := test.TestDB.First(&updatedTask, task.ID).Error
				assert.NoError(t, err)

				assert.Equal(t, "Only Title Updated", updatedTask.Title)
				assert.Equal(t, 30, updatedTask.Points)
				assert.Equal(t, "Updated Description", updatedTask.Description)
				assert.Equal(t, 300.0, updatedTask.Budget)
			},
		},
		{
			name:   "Non-organizer cannot update task",
			userID: otherUser.ID,
			taskID: task.ID,
			request: api.UpdateTaskRequest{
				Title: strPtr("Should Not Update"),
			},
			expectedCode: http.StatusForbidden,
			validateFunc: func(t *testing.T, response api.APIResponse) {
				assert.Contains(t, response.Error, "Only the event organizer can update tasks")
			},
		},
		{
			name:   "Task not found",
			userID: user.ID,
			taskID: 9999, // Non-existent task ID
			request: api.UpdateTaskRequest{
				Title: strPtr("Should Not Update"),
			},
			expectedCode: http.StatusNotFound,
			validateFunc: func(t *testing.T, response api.APIResponse) {
				assert.Contains(t, response.Error, "Task not found")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := test.CreateTestContext(t, tc.userID)

			requestJSON, err := json.Marshal(tc.request)
			assert.NoError(t, err)

			c.Request = httptest.NewRequest("PUT", fmt.Sprintf("/tasks/%d", tc.taskID), bytes.NewBuffer(requestJSON))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = []gin.Param{{Key: "id", Value: fmt.Sprintf("%d", tc.taskID)}}

			handlers.UpdateTask(c, test.TestDB)

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

func TestDeleteTask(t *testing.T) {
	cleanup := test.SetupTestDB(t)
	defer cleanup()

	user := test.CreateTestUser(t)
	event := test.CreateTestEvent(t, user.ID)
	task := test.CreateTestTask(t, event.ID)

	otherUser := test.CreateTestUser(t)
	test.AddEventParticipant(t, event.ID, otherUser.ID)

	testCases := []struct {
		name         string
		userID       uint
		taskID       uint
		expectedCode int
		validateFunc func(t *testing.T, response api.APIResponse)
	}{
		{
			name:         "Task not found",
			userID:       user.ID,
			taskID:       9999, // Non-existent task ID
			expectedCode: http.StatusNotFound,
			validateFunc: func(t *testing.T, response api.APIResponse) {
				assert.Contains(t, response.Error, "Task not found")
			},
		},
		{
			name:         "Non-organizer cannot delete task",
			userID:       otherUser.ID,
			taskID:       task.ID,
			expectedCode: http.StatusForbidden,
			validateFunc: func(t *testing.T, response api.APIResponse) {
				assert.Contains(t, response.Error, "Only the event organizer can delete tasks")

				var count int64
				test.TestDB.Model(&models.Task{}).Where("id = ?", task.ID).Count(&count)
				assert.Equal(t, int64(1), count)
			},
		},
		{
			name:         "Successfully delete task as organizer",
			userID:       user.ID,
			taskID:       task.ID,
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response api.APIResponse) {
				assert.Equal(t, "Task deleted successfully", response.Message)

				var count int64
				test.TestDB.Model(&models.Task{}).Where("id = ?", task.ID).Count(&count)
				assert.Equal(t, int64(0), count)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := test.CreateTestContext(t, tc.userID)

			c.Request = httptest.NewRequest("DELETE", fmt.Sprintf("/tasks/%d", tc.taskID), nil)
			c.Params = []gin.Param{{Key: "id", Value: fmt.Sprintf("%d", tc.taskID)}}

			handlers.DeleteTask(c, test.TestDB)

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

func TestAssignToTask(t *testing.T) {
	cleanup := test.SetupTestDB(t)
	defer cleanup()

	organizer := test.CreateTestUser(t)
	event := test.CreateTestEvent(t, organizer.ID)
	task := test.CreateTestTask(t, event.ID)

	participant := test.CreateTestUser(t)
	test.AddEventParticipant(t, event.ID, participant.ID)

	participantTask := test.CreateTestTask(t, event.ID)
	participantTask.AssignedTo = &participant.ID
	test.TestDB.Save(&participantTask)

	assignedTask := test.CreateTestTask(t, event.ID)
	differentUser := test.CreateTestUser(t)
	assignedTask.AssignedTo = &differentUser.ID
	test.TestDB.Save(&assignedTask)

	testCases := []struct {
		name         string
		userID       uint
		taskID       uint
		expectedCode int
		validateFunc func(t *testing.T, response api.APIResponse)
	}{
		{
			name:         "Task not found",
			userID:       participant.ID,
			taskID:       9999, // Non-existent task ID
			expectedCode: http.StatusNotFound,
			validateFunc: func(t *testing.T, response api.APIResponse) {
				assert.Contains(t, response.Error, "Task not found")
			},
		},
		{
			name:         "Successfully assign task to self",
			userID:       participant.ID,
			taskID:       task.ID,
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response api.APIResponse) {
				assert.Contains(t, response.Message, "You have been assigned to the task")

				var updatedTask models.Task
				err := test.TestDB.First(&updatedTask, task.ID).Error
				assert.NoError(t, err)
				assert.NotNil(t, updatedTask.AssignedTo)
				assert.Equal(t, participant.ID, *updatedTask.AssignedTo)
			},
		},
		{
			name:         "Successfully unassign task from self",
			userID:       participant.ID,
			taskID:       participantTask.ID,
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response api.APIResponse) {
				assert.Contains(t, response.Message, "You have been unassigned from the task")

				var updatedTask models.Task
				err := test.TestDB.First(&updatedTask, participantTask.ID).Error
				assert.NoError(t, err)
				assert.Nil(t, updatedTask.AssignedTo)
			},
		},
		{
			name:         "Cannot assign already assigned task",
			userID:       organizer.ID,
			taskID:       assignedTask.ID, // Task assigned to different user
			expectedCode: http.StatusBadRequest,
			validateFunc: func(t *testing.T, response api.APIResponse) {
				assert.Contains(t, response.Error, "Task already assigned to another user")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := test.CreateTestContext(t, tc.userID)

			c.Request = httptest.NewRequest("PUT", fmt.Sprintf("/tasks/%d/assign", tc.taskID), nil)
			c.Params = []gin.Param{{Key: "id", Value: fmt.Sprintf("%d", tc.taskID)}}

			handlers.AssignToTask(c, test.TestDB)

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

func TestCompleteTask(t *testing.T) {
	cleanup := test.SetupTestDB(t)
	defer cleanup()

	organizer := test.CreateTestUser(t)
	event := test.CreateTestEvent(t, organizer.ID)

	participant1 := test.CreateTestUser(t)
	test.AddEventParticipant(t, event.ID, participant1.ID)

	participant2 := test.CreateTestUser(t)
	test.AddEventParticipant(t, event.ID, participant2.ID)

	unassignedTask := test.CreateTestTask(t, event.ID)

	assignedTask := test.CreateTestTask(t, event.ID)
	assignedTask.AssignedTo = &participant1.ID
	test.TestDB.Save(&assignedTask)

	completedTask := test.CreateTestTask(t, event.ID)
	completedTask.AssignedTo = &participant1.ID
	completedTask.IsCompleted = true
	test.TestDB.Save(&completedTask)

	testCases := []struct {
		name         string
		userID       uint
		taskID       uint
		expectedCode int
		validateFunc func(t *testing.T, response api.APIResponse)
	}{
		{
			name:         "Task not found",
			userID:       participant1.ID,
			taskID:       9999, // Non-existent task ID
			expectedCode: http.StatusNotFound,
			validateFunc: func(t *testing.T, response api.APIResponse) {
				assert.Contains(t, response.Error, "Task not found")
			},
		},
		{
			name:         "Cannot complete unassigned task",
			userID:       participant1.ID,
			taskID:       unassignedTask.ID,
			expectedCode: http.StatusForbidden,
			validateFunc: func(t *testing.T, response api.APIResponse) {
				assert.Contains(t, response.Error, "You can only complete your own tasks")
			},
		},
		{
			name:         "Cannot complete task assigned to another user",
			userID:       participant2.ID,
			taskID:       assignedTask.ID,
			expectedCode: http.StatusForbidden,
			validateFunc: func(t *testing.T, response api.APIResponse) {
				assert.Contains(t, response.Error, "You can only complete your own tasks")
			},
		},
		{
			name:         "Successfully complete assigned task",
			userID:       participant1.ID,
			taskID:       assignedTask.ID,
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response api.APIResponse) {
				var updatedTask models.Task
				err := test.TestDB.First(&updatedTask, assignedTask.ID).Error
				assert.NoError(t, err)
				assert.True(t, updatedTask.IsCompleted)

				var user models.User
				err = test.TestDB.First(&user, participant1.ID).Error
				assert.NoError(t, err)
				assert.Equal(t, assignedTask.Points, user.TotalScore)

				var eventScore models.EventScore
				err = test.TestDB.Where("event_id = ? AND user_id = ?", event.ID, participant1.ID).First(&eventScore).Error
				assert.NoError(t, err)
				assert.Equal(t, float64(assignedTask.Points), eventScore.Score)
			},
		},
		{
			name:         "Successfully uncomplete completed task",
			userID:       participant1.ID,
			taskID:       completedTask.ID,
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response api.APIResponse) {
				var updatedTask models.Task
				err := test.TestDB.First(&updatedTask, completedTask.ID).Error
				assert.NoError(t, err)
				assert.False(t, updatedTask.IsCompleted)

				var user models.User
				err = test.TestDB.First(&user, participant1.ID).Error
				assert.NoError(t, err)
				assert.Equal(t, assignedTask.Points-completedTask.Points, user.TotalScore)

				var eventScore models.EventScore
				err = test.TestDB.Where("event_id = ? AND user_id = ?", event.ID, participant1.ID).First(&eventScore).Error
				assert.NoError(t, err)
				assert.Equal(t, float64(assignedTask.Points-completedTask.Points), eventScore.Score)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := test.CreateTestContext(t, tc.userID)

			c.Request = httptest.NewRequest("PUT", fmt.Sprintf("/tasks/%d/complete", tc.taskID), nil)
			c.Params = []gin.Param{{Key: "id", Value: fmt.Sprintf("%d", tc.taskID)}}

			handlers.CompleteTask(c, test.TestDB)

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
