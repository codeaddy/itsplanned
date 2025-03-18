package handlers

import (
	"itsplanned/auth"
	"itsplanned/models/api"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary Get Google OAuth URL
// @Description Get the URL for Google OAuth authorization
// @Tags calendar
// @Produce json
// @Security BearerAuth
// @Success 200 {object} api.GoogleOAuthURLResponse "OAuth URL generated successfully"
// @Failure 401 {object} api.APIResponse "Unauthorized"
// @Router /auth/google [get]
func GetGoogleOAuthURL(c *gin.Context) {
	url := auth.GetGoogleOAuthURL("randomState")
	c.JSON(http.StatusOK, api.GoogleOAuthURLResponse{URL: url})
}

// @Summary Google OAuth callback
// @Description Handle the callback from Google OAuth and exchange code for tokens
// @Tags calendar
// @Produce json
// @Param code query string true "Authorization code from Google"
// @Success 200 {object} api.GoogleOAuthCallbackResponse "Tokens received successfully"
// @Failure 400 {object} api.APIResponse "Authorization code not provided"
// @Failure 500 {object} api.APIResponse "Failed to get access token"
// @Router /auth/google/callback [get]
func GoogleOAuthCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Authorization code not provided"})
		return
	}

	token, err := auth.ExchangeCodeForToken(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.APIResponse{Error: "Failed to get access token"})
		return
	}

	c.JSON(http.StatusOK, api.GoogleOAuthCallbackResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	})
}
