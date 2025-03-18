package handlers

import (
	"fmt"
	"itsplanned/models"
	"itsplanned/models/api"
	"itsplanned/security"
	"net/http"

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
	if userID != request.UserID {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "You should save only your tokens"})
		return
	}

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
		Expiry:       request.Expiry.String(),
	}

	if err := db.Create(&token).Error; err != nil {
		c.JSON(http.StatusInternalServerError, api.APIResponse{Error: "Failed to save token"})
		return
	}

	c.JSON(http.StatusOK, api.APIResponse{Message: "Token saved successfully"})
}
