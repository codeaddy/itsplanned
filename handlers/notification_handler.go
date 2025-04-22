package handlers

import (
	"itsplanned/models/api"
	"itsplanned/services/notification"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterDeviceToken godoc
// @Summary Register device token
// @Description Register a device token for push notifications
// @Tags notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body api.RegisterDeviceTokenRequest true "Device token registration details"
// @Success 200 {object} api.APIResponse "Device token registered successfully"
// @Failure 400 {object} api.APIResponse "Invalid payload"
// @Failure 401 {object} api.APIResponse "Unauthorized"
// @Failure 500 {object} api.APIResponse "Failed to register device token"
// @Router /notifications/device-token [post]
func RegisterDeviceToken(c *gin.Context, db *gorm.DB) {
	var request api.RegisterDeviceTokenRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Invalid payload"})
		return
	}

	userID, _ := c.Get("user_id")
	userIDUint := userID.(uint)

	err := notification.StoreDeviceToken(userIDUint, request.DeviceToken, request.DeviceType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.APIResponse{Error: "Failed to register device token"})
		return
	}

	c.JSON(http.StatusOK, api.APIResponse{Message: "Device token registered successfully"})
}
