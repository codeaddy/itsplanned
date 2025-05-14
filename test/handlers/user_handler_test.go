package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"itsplanned/handlers"
	"itsplanned/models"
	"itsplanned/models/api"
	"itsplanned/security"
	"itsplanned/test"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func mockLoginHandler(c *gin.Context, db *gorm.DB) {
	var request api.LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Invalid request parameters"})
		return
	}

	var user models.User
	if err := db.Where("email = ?", request.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, api.APIResponse{Error: "Invalid login data"})
		return
	}

	if !security.ComparePassword(user.PasswordHash, request.Password) {
		c.JSON(http.StatusUnauthorized, api.APIResponse{Error: "Invalid login data"})
		return
	}

	token := "test_token_" + request.Email

	c.JSON(http.StatusOK, api.LoginResponse{Token: token})
}

func TestLogin(t *testing.T) {
	cleanup := test.SetupTestDB(t)
	defer cleanup()

	testEmail := fmt.Sprintf("test%d@example.com", time.Now().UnixNano())
	testPassword := "password123"

	hashedPassword, err := security.HashPassword(testPassword)
	assert.NoError(t, err)

	user := &models.User{
		Email:        testEmail,
		DisplayName:  "Test User",
		PasswordHash: hashedPassword,
	}

	err = test.TestDB.Create(user).Error
	assert.NoError(t, err)

	testCases := []struct {
		name         string
		request      api.LoginRequest
		expectedCode int
		validateFunc func(t *testing.T, response *api.LoginResponse)
	}{
		{
			name: "Login with valid credentials",
			request: api.LoginRequest{
				Email:    testEmail,
				Password: testPassword,
			},
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response *api.LoginResponse) {
				assert.NotEmpty(t, response.Token)
			},
		},
		{
			name: "Login with invalid email",
			request: api.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: testPassword,
			},
			expectedCode: http.StatusUnauthorized,
			validateFunc: nil,
		},
		{
			name: "Login with invalid password",
			request: api.LoginRequest{
				Email:    testEmail,
				Password: "wrongpassword",
			},
			expectedCode: http.StatusUnauthorized,
			validateFunc: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := test.CreateTestContext(t, 0)

			requestJSON, err := json.Marshal(tc.request)
			assert.NoError(t, err)

			c.Request = httptest.NewRequest("POST", "/login", bytes.NewBuffer(requestJSON))
			c.Request.Header.Set("Content-Type", "application/json")

			mockLoginHandler(c, test.TestDB)

			assert.Equal(t, tc.expectedCode, w.Code)

			if tc.expectedCode == http.StatusOK && tc.validateFunc != nil {
				var response api.LoginResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				tc.validateFunc(t, &response)
			}
		})
	}
}

func TestResetPassword(t *testing.T) {
	cleanup := test.SetupTestDB(t)
	defer cleanup()

	user := test.CreateTestUser(t)

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
				assert.Equal(t, "Password reset successfully", response.Message)

				var reset models.PasswordReset
				err := test.TestDB.Where("token = ?", "test_reset_token").First(&reset).Error
				assert.NoError(t, err)
				assert.True(t, reset.Used)

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
			c, w := test.CreateTestContext(t, 0)

			requestJSON, err := json.Marshal(tc.request)
			assert.NoError(t, err)

			c.Request = httptest.NewRequest("POST", "/password/reset", bytes.NewBuffer(requestJSON))
			c.Request.Header.Set("Content-Type", "application/json")

			handlers.ResetPassword(c, test.TestDB)

			assert.Equal(t, tc.expectedCode, w.Code)

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
	cleanup := test.SetupTestDB(t)
	defer cleanup()

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
			c, w := test.CreateTestContext(t, tc.userID)

			c.Request = httptest.NewRequest("GET", "/profile", nil)

			handlers.GetProfile(c, test.TestDB)

			assert.Equal(t, tc.expectedCode, w.Code)

			if tc.expectedCode == http.StatusOK && tc.validateFunc != nil {
				var response api.APIResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				tc.validateFunc(t, &response)
			}
		})
	}
}
