package scheduler

import (
	"context"
	"fmt"
	"itsplanned/auth"
	"itsplanned/handlers"
	"itsplanned/models"
	"itsplanned/security"
	"log"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

type Task struct {
	Interval time.Duration
	Func     func()
}

type Scheduler struct {
	tasks []Task
	db    *gorm.DB
	done  chan bool
}

func NewScheduler(db *gorm.DB) *Scheduler {
	return &Scheduler{
		tasks: []Task{},
		db:    db,
		done:  make(chan bool),
	}
}

func (s *Scheduler) AddTask(interval time.Duration, f func()) {
	s.tasks = append(s.tasks, Task{
		Interval: interval,
		Func:     f,
	})
}

func (s *Scheduler) Start() {
	for i := range s.tasks {
		task := s.tasks[i]
		go func() {
			ticker := time.NewTicker(task.Interval)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					task.Func()
				case <-s.done:
					return
				}
			}
		}()
	}
}

func (s *Scheduler) Stop() {
	close(s.done)
}

func (s *Scheduler) SetupCalendarSyncTask() {
	s.AddTask(15*time.Minute, func() {
		log.Println("Starting calendar events sync task...")
		s.SyncCalendarEvents()
	})
}

func (s *Scheduler) SyncCalendarEvents() {
	var tokens []models.UserToken
	if err := s.db.Find(&tokens).Error; err != nil {
		log.Printf("Error fetching user tokens: %v", err)
		return
	}

	log.Printf("Starting calendar sync for %d users with tokens", len(tokens))

	for _, token := range tokens {
		log.Printf("Processing token for user ID: %d, expiry: %s", token.UserID, token.Expiry)
		s.processUserToken(token)
	}
}

func (s *Scheduler) processUserToken(token models.UserToken) {
	accessToken, err := s.refreshTokenIfNeeded(token)
	if err != nil {
		log.Printf("Error refreshing token for user %d: %v", token.UserID, err)
		return
	}

	err = handlers.ImportCalendarEventsForUser(s.db, token.UserID, accessToken)
	if err != nil {
		log.Printf("Error importing calendar events for user %d: %v", token.UserID, err)
	}
}

func (s *Scheduler) refreshTokenIfNeeded(token models.UserToken) (string, error) {
	accessToken, err := security.DecryptToken(token.AccessToken)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt access token: %v", err)
	}

	if token.RefreshToken == "" {
		return accessToken, nil
	}

	refreshToken, err := security.DecryptToken(token.RefreshToken)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt refresh token: %v", err)
	}

	expiryTime, err := time.Parse(time.RFC3339, token.Expiry)
	if err != nil {
		return "", fmt.Errorf("failed to parse token expiry: %v", err)
	}

	if time.Until(expiryTime) > 5*time.Minute {
		return accessToken, nil
	}

	ctx := context.Background()
	config := &oauth2.Config{
		ClientID:     auth.GoogleOAuthConfig.ClientID,
		ClientSecret: auth.GoogleOAuthConfig.ClientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  auth.GoogleOAuthConfig.RedirectURL,
		Scopes:       []string{"https://www.googleapis.com/auth/calendar.readonly"},
	}

	oldToken := &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Expiry:       expiryTime,
	}

	tokenSource := config.TokenSource(ctx, oldToken)
	newToken, err := tokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("failed to refresh token: %v", err)
	}

	newExpiryStr := newToken.Expiry.Format(time.RFC3339)

	newAccessToken, err := security.EncryptToken(newToken.AccessToken)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt new access token: %v", err)
	}

	var newRefreshToken string
	if newToken.RefreshToken != "" {
		newRefreshToken, err = security.EncryptToken(newToken.RefreshToken)
		if err != nil {
			return "", fmt.Errorf("failed to encrypt new refresh token: %v", err)
		}
	} else {
		newRefreshToken = token.RefreshToken
	}

	updates := map[string]interface{}{
		"access_token": newAccessToken,
		"expiry":       newExpiryStr,
	}

	if newToken.RefreshToken != "" {
		updates["refresh_token"] = newRefreshToken
	}

	if err := s.db.Model(&models.UserToken{}).Where("id = ?", token.ID).Updates(updates).Error; err != nil {
		return "", fmt.Errorf("failed to update token in database: %v", err)
	}

	return newToken.AccessToken, nil
}
