package handlers_test

import (
	"bytes"
	"encoding/json"
	"itsplanned/handlers"
	"itsplanned/models"
	"itsplanned/models/api"
	"itsplanned/test"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	// Setup test database
	cleanup := test.SetupTestDB(t)
	defer cleanup()

	testCases := []struct {
		name         string
		request      api.RegisterRequest
		expectedCode int
		validateFunc func(t *testing.T, response *api.APIResponse)
	}{
		{
			name: "Register new user",
			request: api.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response *api.APIResponse) {
				// Should have success message
				assert.Equal(t, "User registered", response.Message)

				// Should have user data
				assert.NotNil(t, response.User)
				assert.Equal(t, "test@example.com", response.User.Email)
				assert.Equal(t, "New User", response.User.DisplayName)

				// Verify user was created in database
				var user models.User
				err := test.TestDB.Where("email = ?", "test@example.com").First(&user).Error
				assert.NoError(t, err)
				assert.NotEmpty(t, user.PasswordHash)
			},
		},
		{
			name: "Register with invalid email",
			request: api.RegisterRequest{
				Email:    "invalid-email",
				Password: "password123",
			},
			expectedCode: http.StatusBadRequest,
			validateFunc: nil,
		},
		{
			name: "Register with short password",
			request: api.RegisterRequest{
				Email:    "test@example.com",
				Password: "short",
			},
			expectedCode: http.StatusBadRequest,
			validateFunc: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test context
			c, w := test.CreateTestContext(t, 0)

			// Convert request to JSON
			requestJSON, err := json.Marshal(tc.request)
			assert.NoError(t, err)

			// Create request
			c.Request = httptest.NewRequest("POST", "/register", bytes.NewBuffer(requestJSON))
			c.Request.Header.Set("Content-Type", "application/json")

			// Call handler
			handlers.Register(c, test.TestDB)

			// Check status code
			assert.Equal(t, tc.expectedCode, w.Code)

			// If we expect a successful response, validate it
			if tc.expectedCode == http.StatusOK && tc.validateFunc != nil {
				var response api.APIResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				tc.validateFunc(t, &response)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	// Setup test database
	cleanup := test.SetupTestDB(t)
	defer cleanup()

	// Create test user
	user := test.CreateTestUser(t)

	testCases := []struct {
		name         string
		request      api.LoginRequest
		expectedCode int
		validateFunc func(t *testing.T, response *api.LoginResponse)
	}{
		{
			name: "Login with valid credentials",
			request: api.LoginRequest{
				Email:    user.Email,
				Password: "password123",
			},
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response *api.LoginResponse) {
				// Should have token
				assert.NotEmpty(t, response.Token)
			},
		},
		{
			name: "Login with invalid email",
			request: api.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			expectedCode: http.StatusUnauthorized,
			validateFunc: nil,
		},
		{
			name: "Login with invalid password",
			request: api.LoginRequest{
				Email:    user.Email,
				Password: "wrongpassword",
			},
			expectedCode: http.StatusUnauthorized,
			validateFunc: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test context
			c, w := test.CreateTestContext(t, 0)

			// Convert request to JSON
			requestJSON, err := json.Marshal(tc.request)
			assert.NoError(t, err)

			// Create request
			c.Request = httptest.NewRequest("POST", "/login", bytes.NewBuffer(requestJSON))
			c.Request.Header.Set("Content-Type", "application/json")

			// Call handler
			handlers.Login(c, test.TestDB)

			// Check status code
			assert.Equal(t, tc.expectedCode, w.Code)

			// If we expect a successful response, validate it
			if tc.expectedCode == http.StatusOK && tc.validateFunc != nil {
				var response api.LoginResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				tc.validateFunc(t, &response)
			}
		})
	}
}

func TestRequestPasswordReset(t *testing.T) {
	// Setup test database
	cleanup := test.SetupTestDB(t)
	defer cleanup()

	// Create test user
	user := test.CreateTestUser(t)

	testCases := []struct {
		name         string
		request      api.PasswordResetRequest
		expectedCode int
		validateFunc func(t *testing.T, response *api.PasswordResetResponse)
	}{
		{
			name: "Request password reset for existing user",
			request: api.PasswordResetRequest{
				Email: user.Email,
			},
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response *api.PasswordResetResponse) {
				// Should have success message
				assert.Equal(t, "If the email exists, a reset link will be sent", response.Message)

				// Verify reset token was created
				var reset models.PasswordReset
				err := test.TestDB.Where("user_id = ?", user.ID).First(&reset).Error
				assert.NoError(t, err)
				assert.NotEmpty(t, reset.Token)
				assert.False(t, reset.Used)
				assert.True(t, reset.ExpiryTime.After(time.Now()))
			},
		},
		{
			name: "Request password reset for non-existent user",
			request: api.PasswordResetRequest{
				Email: "nonexistent@example.com",
			},
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response *api.PasswordResetResponse) {
				// Should have success message
				assert.Equal(t, "If the email exists, a reset link will be sent", response.Message)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test context
			c, w := test.CreateTestContext(t, 0)

			// Convert request to JSON
			requestJSON, err := json.Marshal(tc.request)
			assert.NoError(t, err)

			// Create request
			c.Request = httptest.NewRequest("POST", "/password/reset-request", bytes.NewBuffer(requestJSON))
			c.Request.Header.Set("Content-Type", "application/json")

			// Call handler
			handlers.RequestPasswordReset(c, test.TestDB)

			// Check status code
			assert.Equal(t, tc.expectedCode, w.Code)

			// If we expect a successful response, validate it
			if tc.expectedCode == http.StatusOK && tc.validateFunc != nil {
				var response api.PasswordResetResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				tc.validateFunc(t, &response)
			}
		})
	}
}

func TestResetPassword(t *testing.T) {
	// Setup test database
	cleanup := test.SetupTestDB(t)
	defer cleanup()

	// Create test user
	user := test.CreateTestUser(t)

	// Create test reset token
	resetToken := models.PasswordReset{
		UserID:     user.ID,
		Token:      "test_reset_token",
		ExpiryTime: time.Now().Add(15 * time.Minute),
		Used:       false,
	}
	err := test.TestDB.Create(&resetToken).Error
	assert.NoError(t, err)

	testCases := []struct {
		name         string
		request      api.ResetPasswordRequest
		expectedCode int
		validateFunc func(t *testing.T, response *api.APIResponse)
	}{
		{
			name: "Reset password with valid token",
			request: api.ResetPasswordRequest{
				Token:       "test_reset_token",
				NewPassword: "newpassword123",
			},
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response *api.APIResponse) {
				// Should have success message
				assert.Equal(t, "Password reset successfully", response.Message)

				// Verify token was marked as used
				var reset models.PasswordReset
				err := test.TestDB.Where("token = ?", "test_reset_token").First(&reset).Error
				assert.NoError(t, err)
				assert.True(t, reset.Used)

				// Verify password was updated
				var updatedUser models.User
				err = test.TestDB.First(&updatedUser, user.ID).Error
				assert.NoError(t, err)
				assert.NotEqual(t, user.PasswordHash, updatedUser.PasswordHash)
			},
		},
		{
			name: "Reset password with invalid token",
			request: api.ResetPasswordRequest{
				Token:       "invalid_token",
				NewPassword: "newpassword123",
			},
			expectedCode: http.StatusBadRequest,
			validateFunc: nil,
		},
		{
			name: "Reset password with expired token",
			request: api.ResetPasswordRequest{
				Token:       "test_reset_token",
				NewPassword: "newpassword123",
			},
			expectedCode: http.StatusBadRequest,
			validateFunc: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test context
			c, w := test.CreateTestContext(t, 0)

			// Convert request to JSON
			requestJSON, err := json.Marshal(tc.request)
			assert.NoError(t, err)

			// Create request
			c.Request = httptest.NewRequest("POST", "/password/reset", bytes.NewBuffer(requestJSON))
			c.Request.Header.Set("Content-Type", "application/json")

			// Call handler
			handlers.ResetPassword(c, test.TestDB)

			// Check status code
			assert.Equal(t, tc.expectedCode, w.Code)

			// If we expect a successful response, validate it
			if tc.expectedCode == http.StatusOK && tc.validateFunc != nil {
				var response api.APIResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				tc.validateFunc(t, &response)
			}
		})
	}
}

func TestUpdateProfile(t *testing.T) {
	// Setup test database
	cleanup := test.SetupTestDB(t)
	defer cleanup()

	// Create test user
	user := test.CreateTestUser(t)

	testCases := []struct {
		name         string
		userID       uint
		request      api.ProfileUpdateRequest
		expectedCode int
		validateFunc func(t *testing.T, response *api.APIResponse)
	}{
		{
			name:   "Update profile with all fields",
			userID: user.ID,
			request: api.ProfileUpdateRequest{
				DisplayName: stringPtr("New Display Name"),
				Bio:         stringPtr("New Bio"),
				Avatar:      stringPtr("https://example.com/avatar.jpg"),
			},
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response *api.APIResponse) {
				// Should have success message
				assert.Equal(t, "Profile updated successfully", response.Message)

				// Should have updated user data
				assert.NotNil(t, response.User)
				assert.Equal(t, "New Display Name", response.User.DisplayName)
				assert.Equal(t, "New Bio", response.User.Bio)
				assert.Equal(t, "https://example.com/avatar.jpg", response.User.Avatar)

				// Verify user was updated in database
				var updatedUser models.User
				err := test.TestDB.First(&updatedUser, user.ID).Error
				assert.NoError(t, err)
				assert.Equal(t, "New Display Name", updatedUser.DisplayName)
				assert.Equal(t, "New Bio", updatedUser.Bio)
				assert.Equal(t, "https://example.com/avatar.jpg", updatedUser.Avatar)
			},
		},
		{
			name:   "Update profile with partial fields",
			userID: user.ID,
			request: api.ProfileUpdateRequest{
				DisplayName: stringPtr("New Display Name"),
			},
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response *api.APIResponse) {
				// Should have success message
				assert.Equal(t, "Profile updated successfully", response.Message)

				// Should have updated user data
				assert.NotNil(t, response.User)
				assert.Equal(t, "New Display Name", response.User.DisplayName)
				assert.Equal(t, user.Bio, response.User.Bio)
				assert.Equal(t, user.Avatar, response.User.Avatar)
			},
		},
		{
			name:         "Update profile with invalid user ID",
			userID:       9999,
			request:      api.ProfileUpdateRequest{},
			expectedCode: http.StatusNotFound,
			validateFunc: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test context
			c, w := test.CreateTestContext(t, tc.userID)

			// Convert request to JSON
			requestJSON, err := json.Marshal(tc.request)
			assert.NoError(t, err)

			// Create request
			c.Request = httptest.NewRequest("PUT", "/profile", bytes.NewBuffer(requestJSON))
			c.Request.Header.Set("Content-Type", "application/json")

			// Call handler
			handlers.UpdateProfile(c, test.TestDB)

			// Check status code
			assert.Equal(t, tc.expectedCode, w.Code)

			// If we expect a successful response, validate it
			if tc.expectedCode == http.StatusOK && tc.validateFunc != nil {
				var response api.APIResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				tc.validateFunc(t, &response)
			}
		})
	}
}

func TestGetProfile(t *testing.T) {
	// Setup test database
	cleanup := test.SetupTestDB(t)
	defer cleanup()

	// Create test user
	user := test.CreateTestUser(t)

	testCases := []struct {
		name         string
		userID       uint
		expectedCode int
		validateFunc func(t *testing.T, response *api.APIResponse)
	}{
		{
			name:         "Get profile for valid user",
			userID:       user.ID,
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response *api.APIResponse) {
				// Should have user data
				assert.NotNil(t, response.User)
				assert.Equal(t, user.Email, response.User.Email)
				assert.Equal(t, user.DisplayName, response.User.DisplayName)
				assert.Equal(t, user.Bio, response.User.Bio)
				assert.Equal(t, user.Avatar, response.User.Avatar)
			},
		},
		{
			name:         "Get profile for invalid user ID",
			userID:       9999,
			expectedCode: http.StatusNotFound,
			validateFunc: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test context
			c, w := test.CreateTestContext(t, tc.userID)

			// Create request
			c.Request = httptest.NewRequest("GET", "/profile", nil)

			// Call handler
			handlers.GetProfile(c, test.TestDB)

			// Check status code
			assert.Equal(t, tc.expectedCode, w.Code)

			// If we expect a successful response, validate it
			if tc.expectedCode == http.StatusOK && tc.validateFunc != nil {
				var response api.APIResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				tc.validateFunc(t, &response)
			}
		})
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
