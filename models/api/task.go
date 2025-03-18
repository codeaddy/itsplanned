package api

// TaskResponse represents a task in the response
type TaskResponse struct {
	ID          uint    `json:"id" example:"1"`
	Title       string  `json:"title" example:"Buy decorations"`
	Description string  `json:"description" example:"Purchase party decorations from the store"`
	Budget      float64 `json:"budget" example:"50.00"`
	Points      int     `json:"points" example:"10"`
	EventID     uint    `json:"event_id" example:"1"`
	AssignedTo  *uint   `json:"assigned_to,omitempty" example:"2"`
	IsCompleted bool    `json:"is_completed" example:"false"`
}

// CreateTaskRequest represents the request to create a new task
type CreateTaskRequest struct {
	Title       string  `json:"title" example:"Buy decorations" binding:"required"`
	Description string  `json:"description" example:"Purchase party decorations from the store"`
	Budget      float64 `json:"budget" example:"50.00"`
	Points      int     `json:"points" example:"10" binding:"required"`
	EventID     uint    `json:"event_id" example:"1" binding:"required"`
	AssignedTo  *uint   `json:"assigned_to,omitempty" example:"2"`
}

// UpdateTaskRequest represents the request to update an existing task
type UpdateTaskRequest struct {
	Title       *string  `json:"title,omitempty" example:"Buy party decorations"`
	Description *string  `json:"description,omitempty" example:"Purchase decorations from the party store"`
	Budget      *float64 `json:"budget,omitempty" example:"60.00"`
	Points      *int     `json:"points,omitempty" example:"15"`
}
