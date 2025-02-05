package handlers

import (
	"fmt"
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
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		Expiry       string `json:"expiry"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	userID, _ := c.Get("user_id")
	if userID != payload.UserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You should save only your tokens"})
		return
	}

	hashedAccessToken, err := security.EncryptToken(payload.AccessToken)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error when encrypting tokens"})
		return
	}
	hashedRefreshToken, err := security.EncryptToken(payload.RefreshToken)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error when encrypting tokens"})
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
		Expiry:       payload.Expiry,
	}

	if err := db.Create(&token).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Token saved successfully"})
}
