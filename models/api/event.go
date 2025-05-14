package api

import "time"

// EventResponse represents an event in API responses
type EventResponse struct {
	ID            uint      `json:"id" example:"1"`
	CreatedAt     time.Time `json:"created_at" example:"2024-03-16T12:00:00Z"`
	UpdatedAt     time.Time `json:"updated_at" example:"2024-03-16T12:00:00Z"`
	Name          string    `json:"name" example:"Birthday Party"`
	Description   string    `json:"description" example:"Celebrating John's 30th birthday"`
	EventDateTime time.Time `json:"event_date_time" example:"2024-04-01T18:00:00Z"`
	InitialBudget float64   `json:"initial_budget" example:"1000.00"`
	OrganizerID   uint      `json:"organizer_id" example:"1"`
	Place         string    `json:"place" example:"Central Park"`
}

// CreateEventRequest represents the request to create a new event
type CreateEventRequest struct {
	Name          string  `json:"name" binding:"required" example:"Birthday Party"`
	Description   string  `json:"description" example:"Celebrating John's 30th birthday"`
	EventDateTime string  `json:"event_date_time" binding:"required" example:"2024-04-01T18:00:00Z"`
	InitialBudget float64 `json:"initial_budget" example:"1000.00"`
	Place         string  `json:"place" example:"Central Park"`
}

// UpdateEventRequest represents the request to update an event
type UpdateEventRequest struct {
	Name          *string  `json:"name,omitempty" example:"Birthday Party"`
	Description   *string  `json:"description,omitempty" example:"Celebrating John's 30th birthday"`
	EventDateTime *string  `json:"event_date_time,omitempty" example:"2024-04-01T18:00:00Z"`
	Budget        *float64 `json:"budget,omitempty" example:"1500.00"`
	Place         *string  `json:"place,omitempty" example:"Central Park"`
}

// EventBudgetResponse represents the response for event budget information
type EventBudgetResponse struct {
	InitialBudget float64 `json:"initial_budget" example:"1000.00"`
	RealBudget    float64 `json:"real_budget" example:"950.00"`
	Difference    float64 `json:"difference" example:"50.00"`
}

// EventLeaderboardEntry represents a single entry in the event leaderboard
type EventLeaderboardEntry struct {
	UserID      uint    `json:"user_id" example:"1"`
	DisplayName string  `json:"display_name" example:"John Doe"`
	Score       float64 `json:"score" example:"85.5"`
	EventID     uint    `json:"event_id" example:"1"`
}

// EventLeaderboardResponse represents the response for event leaderboard
type EventLeaderboardResponse struct {
	Leaderboard []EventLeaderboardEntry `json:"leaderboard"`
}

// EventParticipantsResponse represents the response for event participants
type EventParticipantsResponse struct {
	Participants []string `json:"participants" example:"[\"John Doe\", \"Jane Smith\"]"`
}

// TimeSlotSuggestion represents a suggested time slot for an event
type TimeSlotSuggestion struct {
	Slot      string `json:"slot" example:"2024-04-01 18:00"`
	BusyCount int    `json:"busy_count" example:"2"`
}

// FindBestTimeSlotsRequest represents the request to find the best time slots
type FindBestTimeSlotsRequest struct {
	EventID      uint   `json:"event_id" example:"1"`
	Date         string `json:"date" example:"2024-04-01"`
	DurationMins int64  `json:"duration_mins" example:"120"`
	StartTime    string `json:"start_time" example:"08:00"`
	EndTime      string `json:"end_time" example:"22:00"`
}

// FindBestTimeSlotsResponse represents the response with suggested time slots
type FindBestTimeSlotsResponse struct {
	Suggestions []TimeSlotSuggestion `json:"suggestions"`
}
