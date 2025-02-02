package handlers

import (
	"itsplanned/models"
	"itsplanned/security"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Сохраняем хэшированный токен в базу
func SaveOAuthToken(c *gin.Context, db *gorm.DB) {
	var payload struct {
		UserID       uint   `json:"user_id"`
		Provider     string `json:"provider"` // google или apple
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		Expiry       string `json:"expiry"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	hashedAccessToken := security.HashToken(payload.AccessToken)
	hashedRefreshToken := security.HashToken(payload.RefreshToken)

	token := models.UserToken{
		UserID:       payload.UserID,
		Provider:     payload.Provider,
		AccessToken:  hashedAccessToken,
		RefreshToken: hashedRefreshToken,
		Expiry:       payload.Expiry,
	}

	if err := db.Create(&token).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Token saved successfully"})
}
