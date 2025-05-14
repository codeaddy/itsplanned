package handlers

import (
	"itsplanned/auth"
	"itsplanned/models/api"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

// @Summary Get Google OAuth URL
// @Description Get the URL for Google OAuth authorization
// @Tags calendar
// @Produce json
// @Security BearerAuth
// @Param redirect_uri query string false "Custom redirect URI"
// @Success 200 {object} api.GoogleOAuthURLResponse "OAuth URL generated successfully"
// @Failure 401 {object} api.APIResponse "Unauthorized"
// @Router /auth/google [get]
func GetGoogleOAuthURL(c *gin.Context) {
	redirectURI := c.Query("redirect_uri")
	url := auth.GetGoogleOAuthURL("randomState", redirectURI)
	c.JSON(http.StatusOK, api.GoogleOAuthURLResponse{URL: url})
}

// @Summary Google OAuth callback
// @Description Handle the callback from Google OAuth and exchange code for tokens
// @Tags calendar
// @Produce json
// @Param code query string true "Authorization code from Google"
// @Param redirect_uri query string false "Custom redirect URI"
// @Param app_redirect query string false "Deeplink URI to redirect to after OAuth"
// @Success 200 {object} api.GoogleOAuthCallbackResponse "Tokens received successfully"
// @Success 302 {string} string "Redirect to app with tokens"
// @Failure 400 {object} api.APIResponse "Authorization code not provided"
// @Failure 500 {object} api.APIResponse "Failed to get access token"
// @Router /auth/google/callback [get]
func GoogleOAuthCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Authorization code not provided"})
		return
	}

	redirectURI := c.Query("redirect_uri")
	token, err := auth.ExchangeCodeForToken(code, redirectURI)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.APIResponse{Error: "Failed to get access token"})
		return
	}

	appRedirect := c.Query("app_redirect")
	if appRedirect != "" {
		appURL, err := url.Parse(appRedirect)
		if err != nil {
			c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Invalid app redirect URL"})
			return
		}

		q := appURL.Query()
		q.Add("access_token", token.AccessToken)
		q.Add("refresh_token", token.RefreshToken)
		q.Add("expiry", token.Expiry.Format("2006-01-02T15:04:05Z07:00"))
		appURL.RawQuery = q.Encode()

		c.Redirect(http.StatusFound, appURL.String())
		return
	}

	c.JSON(http.StatusOK, api.GoogleOAuthCallbackResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	})
}

// @Summary OAuth Web to App Redirect
// @Description Redirects from web OAuth flow to mobile app via deeplink
// @Tags calendar
// @Produce html
// @Param code query string true "Authorization code from Google"
// @Param state query string false "State parameter for security"
// @Success 302 {string} string "Redirect to app deeplink"
// @Router /auth/web-to-app [get]
func WebToAppRedirect(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	appURI := "itsplanned://callback/auth"

	appURL, _ := url.Parse(appURI)
	q := appURL.Query()
	q.Add("code", code)
	if state != "" {
		q.Add("state", state)
	}
	appURL.RawQuery = q.Encode()

	c.Redirect(http.StatusFound, appURL.String())
}
