package api

import "time"

// GoogleOAuthURLResponse represents the response containing the Google OAuth URL
type GoogleOAuthURLResponse struct {
	URL string `json:"url" example:"https://accounts.google.com/o/oauth2/auth?..."`
}

// GoogleOAuthCallbackResponse represents the response from the OAuth callback
type GoogleOAuthCallbackResponse struct {
	AccessToken  string    `json:"access_token" example:"ya29.a0AfB_byC..."`
	RefreshToken string    `json:"refresh_token" example:"1//04dK..."`
	Expiry       time.Time `json:"expiry" example:"2024-03-16T15:04:05Z"`
}

// SaveOAuthTokenRequest represents the request to save OAuth tokens
type SaveOAuthTokenRequest struct {
	UserID       uint      `json:"user_id" example:"1" binding:"required"`
	AccessToken  string    `json:"access_token" example:"ya29.a0AfB_byC..." binding:"required"`
	RefreshToken string    `json:"refresh_token" example:"1//04dK..." binding:"required"`
	Expiry       time.Time `json:"expiry" example:"2024-03-16T15:04:05Z" binding:"required"`
}
