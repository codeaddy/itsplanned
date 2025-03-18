package api

// GenerateInviteLinkRequest represents the request to generate an invite link for an event
type GenerateInviteLinkRequest struct {
	EventID uint `json:"event_id" example:"1"`
}

// GenerateInviteLinkResponse represents the response containing the generated invite link
type GenerateInviteLinkResponse struct {
	InviteLink string `json:"invite_link" example:"http://localhost:8080/events/join/abc123"`
}

// JoinEventResponse represents the response when a user successfully joins an event
type JoinEventResponse struct {
	Message string `json:"message" example:"Successfully joined event"`
}
