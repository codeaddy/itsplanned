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

func TestSaveOAuthToken(t *testing.T) {
	// Setup test database
	cleanup := test.SetupTestDB(t)
	defer cleanup()

	// Create test user
	user := test.CreateTestUser(t)

	testCases := []struct {
		name         string
		userID       uint
		request      api.SaveOAuthTokenRequest
		expectedCode int
		validateFunc func(t *testing.T)
	}{
		{
			name:   "Save OAuth token for new user",
			userID: user.ID,
			request: api.SaveOAuthTokenRequest{
				AccessToken:  "test_access_token",
				RefreshToken: "test_refresh_token",
				Expiry:       time.Now().Add(24 * time.Hour),
			},
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T) {
				// Verify token was saved
				var token models.UserToken
				err := test.TestDB.Where("user_id = ?", user.ID).First(&token).Error
				assert.NoError(t, err)
				assert.Equal(t, user.ID, token.UserID)
				assert.NotEmpty(t, token.AccessToken)
				assert.NotEmpty(t, token.RefreshToken)
				assert.NotEmpty(t, token.Expiry)
			},
		},
		{
			name:   "Update existing OAuth token",
			userID: user.ID,
			request: api.SaveOAuthTokenRequest{
				AccessToken:  "new_access_token",
				RefreshToken: "new_refresh_token",
				Expiry:       time.Now().Add(48 * time.Hour),
			},
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T) {
				// Verify token was updated
				var token models.UserToken
				err := test.TestDB.Where("user_id = ?", user.ID).First(&token).Error
				assert.NoError(t, err)
				assert.Equal(t, user.ID, token.UserID)
				assert.NotEmpty(t, token.AccessToken)
				assert.NotEmpty(t, token.RefreshToken)
				assert.NotEmpty(t, token.Expiry)

				// Verify old token was deleted
				var count int64
				test.TestDB.Model(&models.UserToken{}).Where("user_id = ?", user.ID).Count(&count)
				assert.Equal(t, int64(1), count)
			},
		},
		{
			name:   "Save OAuth token with invalid user ID",
			userID: 9999,
			request: api.SaveOAuthTokenRequest{
				AccessToken:  "test_access_token",
				RefreshToken: "test_refresh_token",
				Expiry:       time.Now().Add(24 * time.Hour),
			},
			expectedCode: http.StatusOK, // Still returns 200 as the user ID is valid
			validateFunc: func(t *testing.T) {
				// Verify token was saved
				var token models.UserToken
				err := test.TestDB.Where("user_id = ?", uint(9999)).First(&token).Error
				assert.NoError(t, err)
				assert.Equal(t, uint(9999), token.UserID)
			},
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
			c.Request = httptest.NewRequest("POST", "/auth/oauth/save", bytes.NewBuffer(requestJSON))
			c.Request.Header.Set("Content-Type", "application/json")

			// Call handler
			handlers.SaveOAuthToken(c, test.TestDB)

			// Check status code
			assert.Equal(t, tc.expectedCode, w.Code)

			// If we expect a successful response, validate it
			if tc.expectedCode == http.StatusOK && tc.validateFunc != nil {
				var response api.APIResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Token saved successfully", response.Message)

				tc.validateFunc(t)
			}
		})
	}
}
