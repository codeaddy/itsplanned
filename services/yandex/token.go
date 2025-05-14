package yandex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type TokenResponse struct {
	IamToken  string `json:"iamToken"`
	ExpiresAt string `json:"expiresAt"`
}

type TokenRequest struct {
	YandexPassportOauthToken string `json:"yandexPassportOauthToken"`
}

var (
	currentToken string
	tokenMutex   sync.RWMutex
	lastRefresh  time.Time
)

func GetToken() (string, error) {
	tokenMutex.RLock()
	if currentToken != "" && time.Since(lastRefresh) < 30*time.Minute {
		token := currentToken
		tokenMutex.RUnlock()
		return token, nil
	}
	tokenMutex.RUnlock()

	return refreshToken()
}

func refreshToken() (string, error) {
	tokenMutex.Lock()
	defer tokenMutex.Unlock()

	if currentToken != "" && time.Since(lastRefresh) < 30*time.Minute {
		return currentToken, nil
	}

	oauthToken := os.Getenv("YANDEX_OAUTH_TOKEN")
	if oauthToken == "" {
		return "", fmt.Errorf("YANDEX_OAUTH_TOKEN is not set")
	}

	request := TokenRequest{
		YandexPassportOauthToken: oauthToken,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal token request: %w", err)
	}

	resp, err := http.Post(
		"https://iam.api.cloud.yandex.net/iam/v1/tokens",
		"application/json",
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		return "", fmt.Errorf("failed to request IAM token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("IAM token request failed with status code: %d", resp.StatusCode)
	}

	var tokenResponse TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	currentToken = tokenResponse.IamToken
	lastRefresh = time.Now()

	return currentToken, nil
}

func Init() error {
	_, err := refreshToken()
	return err
}
