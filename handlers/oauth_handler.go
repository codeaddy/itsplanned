package handlers

import (
	"fmt"
	"itsplanned/models"
	"itsplanned/models/api"
	"itsplanned/security"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// @Summary Save OAuth tokens
// @Description Save the OAuth tokens for a user
// @Tags calendar
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body api.SaveOAuthTokenRequest true "OAuth token details"
// @Success 200 {object} api.APIResponse "Token saved successfully"
// @Failure 400 {object} api.APIResponse "Invalid payload or unauthorized token save attempt"
// @Failure 500 {object} api.APIResponse "Failed to save token"
// @Router /auth/oauth/save [post]
func SaveOAuthToken(c *gin.Context, db *gorm.DB) {
	var request api.SaveOAuthTokenRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Invalid payload"})
		return
	}

	userID, _ := c.Get("user_id")

	hashedAccessToken, err := security.EncryptToken(request.AccessToken)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Error when encrypting tokens"})
		return
	}
	hashedRefreshToken, err := security.EncryptToken(request.RefreshToken)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Error when encrypting tokens"})
		return
	}

	var existingToken models.UserToken
	if err := db.Where("user_id = ?", userID).First(&existingToken).Error; err == nil {
		db.Delete(&existingToken)
	}

	token := models.UserToken{
		UserID:       userID.(uint),
		AccessToken:  hashedAccessToken,
		RefreshToken: hashedRefreshToken,
		Expiry:       request.Expiry.Format(time.RFC3339),
	}

	if err := db.Create(&token).Error; err != nil {
		c.JSON(http.StatusInternalServerError, api.APIResponse{Error: "Failed to save token"})
		return
	}

	c.JSON(http.StatusOK, api.APIResponse{Message: "Token saved successfully"})
}

// @Summary Delete OAuth tokens
// @Description Delete the OAuth tokens for a user
// @Tags calendar
// @Produce json
// @Security BearerAuth
// @Success 200 {object} api.APIResponse "Token deleted successfully"
// @Failure 404 {object} api.APIResponse "Token not found"
// @Failure 500 {object} api.APIResponse "Failed to delete token"
// @Router /auth/oauth/delete [delete]
func DeleteOAuthToken(c *gin.Context, db *gorm.DB) {
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
