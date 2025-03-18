package api

import "time"

// CalendarEventResponse represents a calendar event in API responses
type CalendarEventResponse struct {
	ID        uint      `json:"id" example:"1"`
	UserID    uint      `json:"user_id" example:"1"`
	Title     string    `json:"title" example:"Team Meeting"`
	StartTime time.Time `json:"start_time" example:"2024-04-01T10:00:00Z"`
	EndTime   time.Time `json:"end_time" example:"2024-04-01T11:00:00Z"`
}

// ImportCalendarEventsResponse represents the response after importing calendar events
type ImportCalendarEventsResponse struct {
	Message string `json:"message" example:"Events imported successfully"`
}
