package handlers

import (
	"fmt"
	"itsplanned/models"
	"itsplanned/models/api"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func toTaskResponse(task *models.Task) *api.TaskResponse {
	if task == nil {
		return nil
	}
	return &api.TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Budget:      task.Budget,
		Points:      task.Points,
		EventID:     task.EventID,
		AssignedTo:  task.AssignedTo,
		IsCompleted: task.IsCompleted,
	}
}

// @Summary Create a new task
// @Description Create a new task for an event
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body api.CreateTaskRequest true "Task creation details"
// @Success 200 {object} api.APIResponse{data=api.TaskResponse} "Task created successfully"
// @Failure 400 {object} api.APIResponse "Invalid payload"
// @Failure 401 {object} api.APIResponse "Unauthorized"
// @Failure 500 {object} api.APIResponse "Failed to create task"
// @Router /tasks [post]
func CreateTask(c *gin.Context, db *gorm.DB) {
	var request api.CreateTaskRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Invalid payload"})
		return
	}

	task := models.Task{
		Title:       request.Title,
		Description: request.Description,
		Budget:      request.Budget,
		Points:      request.Points,
		EventID:     request.EventID,
		AssignedTo:  request.AssignedTo,
	}

	if err := db.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, api.APIResponse{Error: "Failed to create task"})
		return
	}

	c.JSON(http.StatusOK, api.APIResponse{
		Message: "Task created",
		Data:    toTaskResponse(&task),
	})
}

// @Summary Toggle task assignment
// @Description Assign the authenticated user to an unassigned task or unassign if already assigned
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Task ID"
// @Success 200 {object} api.APIResponse{data=api.TaskResponse} "Task assignment toggled successfully"
// @Failure 400 {object} api.APIResponse "Task already assigned to another user"
// @Failure 401 {object} api.APIResponse "Unauthorized"
// @Failure 404 {object} api.APIResponse "Task not found"
// @Router /tasks/{id}/assign [put]
func AssignToTask(c *gin.Context, db *gorm.DB) {
	var task models.Task
	id := c.Param("id")

	if err := db.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, api.APIResponse{Error: "Task not found"})
		return
	}

	userID, _ := c.Get("user_id")
	userIDUint := userID.(uint)

	// Check if user is already assigned to the task
	if task.AssignedTo != nil && *task.AssignedTo == userIDUint {
		// Unassign the user
		task.AssignedTo = nil
		db.Save(&task)
		c.JSON(http.StatusOK, api.APIResponse{
			Message: "You have been unassigned from the task",
			Data:    toTaskResponse(&task),
		})
		return
	}

	// Check if task is already assigned to someone else
	if task.AssignedTo != nil && *task.AssignedTo != userIDUint {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Task already assigned to another user"})
		return
	}

	// Assign the user
	task.AssignedTo = &userIDUint
	db.Save(&task)

	c.JSON(http.StatusOK, api.APIResponse{
		Message: "You have been assigned to the task",
		Data:    toTaskResponse(&task),
	})
}

// @Summary Toggle task completion
// @Description Mark a task as completed or uncompleted and update user scores accordingly
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Task ID"
// @Success 200 {object} api.APIResponse{data=api.TaskResponse} "Task completion toggled successfully"
// @Failure 401 {object} api.APIResponse "Unauthorized"
// @Failure 403 {object} api.APIResponse "Forbidden - not assigned to the task"
// @Failure 404 {object} api.APIResponse "Task not found"
// @Router /tasks/{id}/complete [put]
func CompleteTask(c *gin.Context, db *gorm.DB) {
	var task models.Task
	id := c.Param("id")

	if err := db.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, api.APIResponse{Error: "Task not found"})
		return
	}

	userID, _ := c.Get("user_id")
	if task.AssignedTo == nil || userID.(uint) != *task.AssignedTo {
		c.JSON(http.StatusForbidden, api.APIResponse{Error: "You can only complete your own tasks"})
		return
	}

	// Toggle completion status
	task.IsCompleted = !task.IsCompleted
	db.Save(&task)

	var user models.User
	if err := db.First(&user, task.AssignedTo).Error; err == nil {
		if task.IsCompleted {
			user.TotalScore += task.Points
		} else {
			user.TotalScore -= task.Points
		}
		db.Save(&user)
	}

	var eventScore models.EventScore
	if err := db.Where("event_id = ? AND user_id = ?", task.EventID, task.AssignedTo).First(&eventScore).Error; err != nil {
		if task.IsCompleted {
			eventScore = models.EventScore{
				EventID: task.EventID,
				UserID:  *task.AssignedTo,
				Score:   float64(task.Points),
			}
			db.Create(&eventScore)
		}
	} else {
		if task.IsCompleted {
			eventScore.Score += float64(task.Points)
		} else {
			eventScore.Score -= float64(task.Points)
		}
		db.Save(&eventScore)
	}

	message := "Task completed"
	if !task.IsCompleted {
		message = "Task uncompleted"
	}

	c.JSON(http.StatusOK, api.APIResponse{
		Message: message,
		Data:    toTaskResponse(&task),
	})
}

// GetTasks godoc
// @Summary Get all tasks for an event
// @Description Get a list of all tasks associated with a specific event
// @Tags tasks
// @Produce json
// @Security BearerAuth
// @Param event_id query int true "Event ID"
// @Success 200 {object} api.APIResponse{data=[]api.TaskResponse} "List of tasks retrieved successfully"
// @Failure 400 {object} api.APIResponse "Invalid event ID"
// @Failure 401 {object} api.APIResponse "Unauthorized"
// @Failure 403 {object} api.APIResponse "Forbidden - not a participant of the event"
// @Failure 404 {object} api.APIResponse "Event not found"
// @Router /tasks [get]
func GetTasks(c *gin.Context, db *gorm.DB) {
	eventIDStr := c.Query("event_id")
	if eventIDStr == "" {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Event ID is required"})
		return
	}

	var eventID uint
	if _, err := fmt.Sscanf(eventIDStr, "%d", &eventID); err != nil {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Invalid event ID format"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, api.APIResponse{Error: "User not authenticated"})
		return
	}

	// Check if the event exists and user is a participant
	var event models.Event
	if err := db.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, api.APIResponse{Error: "Event not found"})
		return
	}

	// Check if user is organizer or participant
	if event.OrganizerID != userID.(uint) {
		var participation models.EventParticipation
		if err := db.Where("event_id = ? AND user_id = ?", eventID, userID).First(&participation).Error; err != nil {
			c.JSON(http.StatusForbidden, api.APIResponse{Error: "You are not a participant of this event"})
			return
		}
	}

	// Get all tasks for the event
	var tasks []models.Task
	if err := db.Where("event_id = ?", eventID).Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, api.APIResponse{Error: "Failed to retrieve tasks"})
		return
	}

	var response []api.TaskResponse
	for _, task := range tasks {
		response = append(response, api.TaskResponse{
			ID:          task.ID,
			Title:       task.Title,
			Description: task.Description,
			Budget:      task.Budget,
			Points:      task.Points,
			EventID:     task.EventID,
			AssignedTo:  task.AssignedTo,
			IsCompleted: task.IsCompleted,
		})
	}

	c.JSON(http.StatusOK, api.APIResponse{Data: response})
}

// GetTask godoc
// @Summary Get task details
// @Description Get detailed information about a specific task
// @Tags tasks
// @Produce json
// @Security BearerAuth
// @Param id path int true "Task ID"
// @Success 200 {object} api.APIResponse{data=api.TaskResponse} "Task details retrieved successfully"
// @Failure 400 {object} api.APIResponse "Invalid task ID"
// @Failure 401 {object} api.APIResponse "Unauthorized"
// @Failure 403 {object} api.APIResponse "Forbidden - not a participant of the event"
// @Failure 404 {object} api.APIResponse "Task not found"
// @Router /tasks/{id} [get]
func GetTask(c *gin.Context, db *gorm.DB) {
	taskIDStr := c.Param("id")
	if taskIDStr == "" {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Task ID is required"})
		return
	}

	var taskID uint
	if _, err := fmt.Sscanf(taskIDStr, "%d", &taskID); err != nil {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Invalid task ID format"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, api.APIResponse{Error: "User not authenticated"})
		return
	}

	// Get the task
	var task models.Task
	if err := db.First(&task, taskID).Error; err != nil {
		c.JSON(http.StatusNotFound, api.APIResponse{Error: "Task not found"})
		return
	}

	// Check if the event exists and user is a participant
	var event models.Event
	if err := db.First(&event, task.EventID).Error; err != nil {
		c.JSON(http.StatusNotFound, api.APIResponse{Error: "Event not found"})
		return
	}

	// Check if user is organizer or participant of the event
	if event.OrganizerID != userID.(uint) {
		var participation models.EventParticipation
		if err := db.Where("event_id = ? AND user_id = ?", task.EventID, userID).First(&participation).Error; err != nil {
			c.JSON(http.StatusForbidden, api.APIResponse{Error: "You are not a participant of this event"})
			return
		}
	}

	response := api.TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Budget:      task.Budget,
		Points:      task.Points,
		EventID:     task.EventID,
		AssignedTo:  task.AssignedTo,
		IsCompleted: task.IsCompleted,
	}

	c.JSON(http.StatusOK, api.APIResponse{Data: response})
}

// @Summary Update task details
// @Description Update details of an existing task
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Task ID"
// @Param request body api.UpdateTaskRequest true "Task update details"
// @Success 200 {object} api.APIResponse{data=api.TaskResponse} "Task updated successfully"
// @Failure 400 {object} api.APIResponse "Invalid payload"
// @Failure 401 {object} api.APIResponse "Unauthorized"
// @Failure 403 {object} api.APIResponse "Forbidden - not a participant of the event or not the organizer"
// @Failure 404 {object} api.APIResponse "Task not found"
// @Router /tasks/{id} [put]
func UpdateTask(c *gin.Context, db *gorm.DB) {
	taskIDStr := c.Param("id")
	if taskIDStr == "" {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Task ID is required"})
		return
	}

	var taskID uint
	if _, err := fmt.Sscanf(taskIDStr, "%d", &taskID); err != nil {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Invalid task ID format"})
		return
	}

	// Get the current user ID from the context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, api.APIResponse{Error: "User not authenticated"})
		return
	}

	// Get the task
	var task models.Task
	if err := db.First(&task, taskID).Error; err != nil {
		c.JSON(http.StatusNotFound, api.APIResponse{Error: "Task not found"})
		return
	}

	// Check if the event exists
	var event models.Event
	if err := db.First(&event, task.EventID).Error; err != nil {
		c.JSON(http.StatusNotFound, api.APIResponse{Error: "Event not found"})
		return
	}

	// Check if the user is the organizer of the event (only organizer can update task details)
	if event.OrganizerID != userID.(uint) {
		c.JSON(http.StatusForbidden, api.APIResponse{Error: "Only the event organizer can update tasks"})
		return
	}

	// Parse the request
	var request api.UpdateTaskRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Invalid payload"})
		return
	}

	// Update task fields if provided
	if request.Title != nil {
		task.Title = *request.Title
	}
	if request.Description != nil {
		task.Description = *request.Description
	}
	if request.Budget != nil {
		task.Budget = *request.Budget
	}
	if request.Points != nil {
		task.Points = *request.Points
	}

	// Save the updated task
	if err := db.Save(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, api.APIResponse{Error: "Failed to update task"})
		return
	}

	c.JSON(http.StatusOK, api.APIResponse{
		Message: "Task updated successfully",
		Data:    toTaskResponse(&task),
	})
}
