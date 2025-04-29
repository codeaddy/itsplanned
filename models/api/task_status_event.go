package api

import "time"

// TaskStatusEventResponse represents a task status change event in API responses
type TaskStatusEventResponse struct {
	ID            uint      `json:"id" example:"1"`
	TaskID        uint      `json:"task_id" example:"5"`
	TaskName      string    `json:"task_name" example:"Buy decorations"`
	OldStatus     string    `json:"old_status,omitempty" example:"unassigned"`
	NewStatus     string    `json:"new_status" example:"assigned"`
	ChangedByID   uint      `json:"changed_by_id" example:"2"`
	ChangedByName string    `json:"changed_by_name" example:"John Doe"`
	IsRead        bool      `json:"is_read" example:"false"`
	EventTime     time.Time `json:"event_time" example:"2024-03-16T12:00:00Z"`
}

// TaskStatusEventsResponse represents a collection of task status events in API responses
type TaskStatusEventsResponse struct {
	Events []TaskStatusEventResponse `json:"events"`
}
