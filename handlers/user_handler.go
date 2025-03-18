package handlers

import (
	"itsplanned/models"
	"itsplanned/models/api"
	"itsplanned/security"
	"itsplanned/services/email"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func toUserResponse(user *models.User) *api.UserResponse {
	if user == nil {
		return nil
	}
	return &api.UserResponse{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Bio:         user.Bio,
		Avatar:      user.Avatar,
	}
}

// @Summary Register a new user
// @Description Register a new user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body api.RegisterRequest true "User registration details"
// @Success 200 {object} api.APIResponse "User registered successfully"
// @Failure 400 {object} api.APIResponse "Invalid payload"
// @Failure 500 {object} api.APIResponse "Failed to hash password"
// @Router /register [post]
func Register(c *gin.Context, db *gorm.DB) {
	var request api.RegisterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Invalid payload"})
		return
	}

	hashedPassword, err := security.HashPassword(request.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.APIResponse{Error: "Failed to hash password"})
		return
	}

	user := models.User{
		Email:        request.Email,
		PasswordHash: hashedPassword,
		DisplayName:  "New User",
	}

	db.Create(&user)
	c.JSON(http.StatusOK, api.APIResponse{Message: "User registered", User: toUserResponse(&user)})
}

// @Summary User login
// @Description Authenticate a user and return a JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body api.LoginRequest true "User login credentials"
// @Success 200 {object} api.LoginResponse "Login successful"
// @Failure 400 {object} api.APIResponse "Invalid payload"
// @Failure 401 {object} api.APIResponse "Invalid credentials"
// @Failure 500 {object} api.APIResponse "Failed to generate token"
// @Router /login [post]
func Login(c *gin.Context, db *gorm.DB) {
	var request api.LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Invalid payload"})
		return
	}

	var user models.User
	if err := db.Where("email = ?", request.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, api.APIResponse{Error: "Invalid credentials"})
		return
	}

	if !security.ComparePassword(user.PasswordHash, request.Password) {
		c.JSON(http.StatusUnauthorized, api.APIResponse{Error: "Invalid credentials"})
		return
	}

	token, err := security.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.APIResponse{Error: "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, api.LoginResponse{Token: token})
}

// @Summary Request password reset
// @Description Request a password reset token for a user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body api.PasswordResetRequest true "User email"
// @Success 200 {object} api.PasswordResetResponse "Reset token generated"
// @Failure 400 {object} api.APIResponse "Invalid payload"
// @Router /password/reset-request [post]
func RequestPasswordReset(c *gin.Context, db *gorm.DB) {
	var request api.PasswordResetRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Invalid payload"})
		return
	}

	var user models.User
	if err := db.Where("email = ?", request.Email).First(&user).Error; err != nil {
		// For security reasons, always return success even if email doesn't exist
		c.JSON(http.StatusOK, api.PasswordResetResponse{
			Message: "If the email exists, a reset link will be sent",
		})
		return
	}

	// Generate reset token
	resetToken := security.GenerateRandomToken()
	expiryTime := time.Now().Add(15 * time.Minute)

	// Save reset token in database
	passwordReset := models.PasswordReset{
		UserID:     user.ID,
		Token:      resetToken,
		ExpiryTime: expiryTime,
	}
	db.Create(&passwordReset)

	// Send email with reset link
	err := email.SendPasswordResetEmail(user.Email, resetToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.APIResponse{
			Error: "Failed to send reset email: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, api.PasswordResetResponse{
		Message: "If the email exists, a reset link will be sent",
	})
}

// @Summary Reset password
// @Description Reset user password using a valid reset token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body api.ResetPasswordRequest true "Reset token and new password"
// @Success 200 {object} api.APIResponse "Password reset successfully"
// @Failure 400 {object} api.APIResponse "Invalid payload or token"
// @Failure 500 {object} api.APIResponse "Internal server error"
// @Router /password/reset [post]
func ResetPassword(c *gin.Context, db *gorm.DB) {
	var request api.ResetPasswordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Invalid payload"})
		return
	}

	var passwordReset models.PasswordReset
	if err := db.Where("token = ? AND used = ? AND expiry_time > ?",
		request.Token, false, time.Now()).First(&passwordReset).Error; err != nil {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Invalid or expired token"})
		return
	}

	// Hash new password
	hashedPassword, err := security.HashPassword(request.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.APIResponse{Error: "Failed to hash password"})
		return
	}

	// Update user password
	var user models.User
	if err := db.First(&user, passwordReset.UserID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, api.APIResponse{Error: "User not found"})
		return
	}

	user.PasswordHash = hashedPassword
	db.Save(&user)

	// Mark reset token as used
	passwordReset.Used = true
	db.Save(&passwordReset)

	c.JSON(http.StatusOK, api.APIResponse{Message: "Password reset successfully"})
}

// @Summary Update user profile
// @Description Update the authenticated user's profile information
// @Tags profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body api.ProfileUpdateRequest true "Profile update data"
// @Success 200 {object} api.APIResponse "Profile updated successfully"
// @Failure 400 {object} api.APIResponse "Invalid payload"
// @Failure 401 {object} api.APIResponse "Unauthorized"
// @Failure 404 {object} api.APIResponse "User not found"
// @Router /profile [put]
func UpdateProfile(c *gin.Context, db *gorm.DB) {
	userID, _ := c.Get("user_id")

	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, api.APIResponse{Error: "User not found"})
		return
	}

	var request api.ProfileUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Invalid payload"})
		return
	}

	if request.DisplayName != nil {
		user.DisplayName = *request.DisplayName
	}
	if request.Bio != nil {
		user.Bio = *request.Bio
	}
	if request.Avatar != nil {
		user.Avatar = *request.Avatar
	}

	db.Save(&user)
	c.JSON(http.StatusOK, api.APIResponse{Message: "Profile updated successfully", User: toUserResponse(&user)})
}

// @Summary Get user profile
// @Description Get the authenticated user's profile information
// @Tags profile
// @Produce json
// @Security BearerAuth
// @Success 200 {object} api.APIResponse "User profile"
// @Failure 401 {object} api.APIResponse "Unauthorized"
// @Failure 404 {object} api.APIResponse "User not found"
// @Router /profile [get]
func GetProfile(c *gin.Context, db *gorm.DB) {
	userID, _ := c.Get("user_id")

	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, api.APIResponse{Error: "User not found"})
		return
	}

	c.JSON(http.StatusOK, api.APIResponse{User: toUserResponse(&user)})
}

// @Summary User logout
// @Description Logout the current user (client-side token removal)
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} api.APIResponse "Logged out successfully"
// @Router /logout [post]
func Logout(c *gin.Context, db *gorm.DB) {
	c.JSON(http.StatusOK, api.APIResponse{Message: "Logged out successfully"})
}
