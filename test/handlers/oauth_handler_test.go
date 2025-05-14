package handlers_test

import (
	"bytes"
	"encoding/json"
	"itsplanned/models"
	"itsplanned/models/api"
	"itsplanned/test"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func mockSaveOAuthToken(c *gin.Context, db *gorm.DB) {
	var request api.SaveOAuthTokenRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Invalid payload"})
		return
	}

	userID, _ := c.Get("user_id")

	var existingToken models.UserToken
	if err := db.Where("user_id = ?", userID).First(&existingToken).Error; err == nil {
		db.Delete(&existingToken)
	}

	token := models.UserToken{
		UserID:       userID.(uint),
		AccessToken:  request.AccessToken,
		RefreshToken: request.RefreshToken,
		Expiry:       request.Expiry.Format(time.RFC3339),
	}

	if err := db.Create(&token).Error; err != nil {
		c.JSON(http.StatusInternalServerError, api.APIResponse{Error: "Failed to save token"})
		return
	}

	c.JSON(http.StatusOK, api.APIResponse{Message: "Token saved successfully"})
}

func TestSaveOAuthToken(t *testing.T) {
	cleanup := test.SetupTestDB(t)
	defer cleanup()

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
				var token models.UserToken
				err := test.TestDB.Where("user_id = ?", user.ID).First(&token).Error
				assert.NoError(t, err)
				assert.Equal(t, user.ID, token.UserID)
				assert.Equal(t, "test_access_token", token.AccessToken)
				assert.Equal(t, "test_refresh_token", token.RefreshToken)
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
				var token models.UserToken
				err := test.TestDB.Where("user_id = ?", user.ID).First(&token).Error
				assert.NoError(t, err)
				assert.Equal(t, user.ID, token.UserID)
				assert.Equal(t, "new_access_token", token.AccessToken)
				assert.Equal(t, "new_refresh_token", token.RefreshToken)
				assert.NotEmpty(t, token.Expiry)

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
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T) {
				var token models.UserToken
				err := test.TestDB.Where("user_id = ?", uint(9999)).First(&token).Error
				assert.NoError(t, err)
				assert.Equal(t, uint(9999), token.UserID)
				assert.Equal(t, "test_access_token", token.AccessToken)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := test.CreateTestContext(t, tc.userID)

			requestJSON, err := json.Marshal(tc.request)
			assert.NoError(t, err)

			c.Request = httptest.NewRequest("POST", "/auth/oauth/save", bytes.NewBuffer(requestJSON))
			c.Request.Header.Set("Content-Type", "application/json")

			mockSaveOAuthToken(c, test.TestDB)

			assert.Equal(t, tc.expectedCode, w.Code)

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

func mockDeleteOAuthToken(c *gin.Context, db *gorm.DB) {
	userID, _ := c.Get("user_id")

	var existingToken models.UserToken
	if err := db.Where("user_id = ?", userID).First(&existingToken).Error; err != nil {
		c.JSON(http.StatusNotFound, api.APIResponse{Error: "OAuth token not found"})
		return
	}

	if err := db.Delete(&existingToken).Error; err != nil {
		c.JSON(http.StatusInternalServerError, api.APIResponse{Error: "Failed to delete token"})
		return
	}

	c.JSON(http.StatusOK, api.APIResponse{Message: "Token deleted successfully"})
}

func TestDeleteOAuthToken(t *testing.T) {
	cleanup := test.SetupTestDB(t)
	defer cleanup()

	user := test.CreateTestUser(t)

	testCases := []struct {
		name         string
		userID       uint
		setupFunc    func()
		expectedCode int
		expectedMsg  string
	}{
		{
			name:   "Delete existing OAuth token",
			userID: user.ID,
			setupFunc: func() {
				token := models.UserToken{
					UserID:       user.ID,
					AccessToken:  "test_access_token",
					RefreshToken: "test_refresh_token",
					Expiry:       time.Now().Add(24 * time.Hour).Format(time.RFC3339),
				}
				test.TestDB.Create(&token)
			},
			expectedCode: http.StatusOK,
			expectedMsg:  "Token deleted successfully",
		},
		{
			name:         "Delete non-existent OAuth token",
			userID:       9999,
			setupFunc:    func() {},
			expectedCode: http.StatusNotFound,
			expectedMsg:  "OAuth token not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test data
			tc.setupFunc()

			c, w := test.CreateTestContext(t, tc.userID)
			c.Request = httptest.NewRequest("DELETE", "/auth/oauth/delete", nil)

			mockDeleteOAuthToken(c, test.TestDB)

			assert.Equal(t, tc.expectedCode, w.Code)

			var response api.APIResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tc.expectedCode == http.StatusOK {
				assert.Equal(t, tc.expectedMsg, response.Message)

				// Verify token is deleted
				var count int64
				test.TestDB.Model(&models.UserToken{}).Where("user_id = ?", tc.userID).Count(&count)
				assert.Equal(t, int64(0), count)
			} else {
				assert.Equal(t, tc.expectedMsg, response.Error)
			}
		})
	}
}
