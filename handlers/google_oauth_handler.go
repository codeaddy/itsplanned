package handlers

import (
	"itsplanned/auth"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Получение ссылки на авторизацию пользователя в Google аккаунте
func GetGoogleOAuthURL(c *gin.Context) {
	url := auth.GetGoogleOAuthURL("randomState")
	c.JSON(http.StatusOK, gin.H{"url": url})
}

// Обмениваем полученный код на токен
func GoogleOAuthCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code not provided"})
		return
	}

	token, err := auth.ExchangeCodeForToken(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get access token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": token.AccessToken, "refresh_token": token.RefreshToken, "expiry": token.Expiry})
}
