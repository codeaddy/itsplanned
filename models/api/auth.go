package api

import "time"

// UserResponse represents the user data in API responses
type UserResponse struct {
	ID          uint      `json:"id" example:"1"`
	CreatedAt   time.Time `json:"created_at" example:"2024-03-16T12:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2024-03-16T12:00:00Z"`
	Email       string    `json:"email" example:"user@example.com"`
	DisplayName string    `json:"display_name" example:"John Doe"`
	Bio         string    `json:"bio,omitempty" example:"Software developer and tech enthusiast"`
	Avatar      string    `json:"avatar,omitempty" example:"https://example.com/avatar.jpg"`
}

// RegisterRequest represents the user registration request
type RegisterRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"secretpassword123"`
}

// LoginRequest represents the user login request
type LoginRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"secretpassword123"`
}

// LoginResponse represents the response after successful login
type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// PasswordResetRequest represents the password reset request
type PasswordResetRequest struct {
	Email string `json:"email" example:"user@example.com"`
}

// PasswordResetResponse represents the response to a password reset request
type PasswordResetResponse struct {
	Message    string `json:"message" example:"If the email exists, a reset link will be sent"`
	ResetToken string `json:"reset_token,omitempty" example:"abc123def456"`
}

// ResetPasswordRequest represents the request to reset password with token
type ResetPasswordRequest struct {
	Token       string `json:"token" example:"abc123def456"`
	NewPassword string `json:"new_password" example:"newpassword123"`
}

// ProfileUpdateRequest represents the request to update user profile
type ProfileUpdateRequest struct {
	DisplayName *string `json:"display_name,omitempty" example:"John Doe"`
	Bio         *string `json:"bio,omitempty" example:"Software developer and tech enthusiast"`
	Avatar      *string `json:"avatar,omitempty" example:"https://example.com/avatar.jpg"`
}

// APIResponse represents a generic API response
type APIResponse struct {
	Message string        `json:"message,omitempty" example:"Operation successful"`
	Error   string        `json:"error,omitempty" example:"Invalid input"`
	User    *UserResponse `json:"user,omitempty"`
	Data    interface{}   `json:"data,omitempty" swaggertype:"object"`
}
