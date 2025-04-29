package auth

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Config for OAuth 2.0
var GoogleOAuthConfig = &oauth2.Config{
	RedirectURL:  "http://localhost:8080/auth/web-to-app",
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	Scopes:       []string{"https://www.googleapis.com/auth/calendar.readonly"},
	Endpoint:     google.Endpoint,
}

// Google OAuth URL generation
func GetGoogleOAuthURL(state string, redirectURI string) string {
	config := &oauth2.Config{
		ClientID:     GoogleOAuthConfig.ClientID,
		ClientSecret: GoogleOAuthConfig.ClientSecret,
		Endpoint:     GoogleOAuthConfig.Endpoint,
		RedirectURL:  GoogleOAuthConfig.RedirectURL,
		Scopes:       GoogleOAuthConfig.Scopes,
	}

	if redirectURI != "" {
		config.RedirectURL = redirectURI
	}

	return config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// Exchange auth code on access_token
func ExchangeCodeForToken(code string, redirectURI string) (*oauth2.Token, error) {
	ctx := context.Background()

	config := &oauth2.Config{
		ClientID:     GoogleOAuthConfig.ClientID,
		ClientSecret: GoogleOAuthConfig.ClientSecret,
		Endpoint:     GoogleOAuthConfig.Endpoint,
		RedirectURL:  GoogleOAuthConfig.RedirectURL,
		Scopes:       GoogleOAuthConfig.Scopes,
	}

	if redirectURI != "" {
		config.RedirectURL = redirectURI
	}

	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %v", err)
	}
	return token, nil
}
