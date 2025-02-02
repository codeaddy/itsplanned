package handlers

import (
	"itsplanned/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ImportCalendarEvents(c *gin.Context, db *gorm.DB) {
	var payload struct {
		Provider     string `json:"provider"` // google или apple
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		UserID       uint   `json:"user_id"`
		Expiry       string `json:"expiry"` // Дата истечения access_token
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	token := models.UserToken{
		UserID:       payload.UserID,
		Provider:     payload.Provider,
		AccessToken:  payload.AccessToken,
		RefreshToken: payload.RefreshToken,
		Expiry:       payload.Expiry,
	}

	if err := db.Create(&token).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Token saved successfully"})
}
