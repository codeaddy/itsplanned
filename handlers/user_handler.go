package handlers

import (
	"itsplanned/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Register(c *gin.Context, db *gorm.DB) {
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	// TODO: добавить хэширование
	passwordHash := "HASH" + payload.Password

	user := models.User{
		Email:        payload.Email,
		PasswordHash: passwordHash,
		DisplayName:  "New User",
	}

	db.Create(&user)
	c.JSON(http.StatusOK, gin.H{"message": "User registered", "user": user})
}

func Login(c *gin.Context, db *gorm.DB) {
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	var user models.User
	if err := db.Where("email = ?", payload.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// TODO: проверять хэш
	if user.PasswordHash != "HASH"+payload.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged in", "user": user})
}

func GetProfile(c *gin.Context, db *gorm.DB) {
	// TODO
}
