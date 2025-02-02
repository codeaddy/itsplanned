package auth

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Конфигурация для OAuth 2.0
var GoogleOAuthConfig = &oauth2.Config{
	RedirectURL:  "http://localhost:8080/auth/google/callback",
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	Scopes:       []string{"https://www.googleapis.com/auth/calendar.readonly"},
	Endpoint:     google.Endpoint,
}

// Генерация ссылки для авторизации пользователя через Google OAuth
func GetGoogleOAuthURL(state string) string {
	return GoogleOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// Обмен кода авторизации на access_token
func ExchangeCodeForToken(code string) (*oauth2.Token, error) {
	ctx := context.Background()
	token, err := GoogleOAuthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %v", err)
	}
	return token, nil
}
