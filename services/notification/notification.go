package notification

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"itsplanned/models"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

// NotificationConfig holds the Apple Push Notification Service (APNS) configuration
type NotificationConfig struct {
	APNSAuthKey   string // Path to .p8 auth key file
	APNSKeyID     string // Key ID from Apple Developer account
	APNSTeamID    string // Team ID from Apple Developer account
	APNSBundleID  string // Bundle ID of your iOS app
	APNSIsSandbox bool   // Whether to use development or production APNS
}

var config NotificationConfig
var db *gorm.DB

// APNS endpoints
const (
	APNSProductionURL = "https://api.push.apple.com/3/device"
	APNSSandboxURL    = "https://api.development.push.apple.com/3/device"
)

// Init initializes the notification service
func Init(database *gorm.DB) error {
	db = database

	config = NotificationConfig{
		APNSAuthKey:   os.Getenv("APNS_AUTH_KEY"),
		APNSKeyID:     os.Getenv("APNS_KEY_ID"),
		APNSTeamID:    os.Getenv("APNS_TEAM_ID"),
		APNSBundleID:  os.Getenv("APNS_BUNDLE_ID"),
		APNSIsSandbox: os.Getenv("APNS_IS_SANDBOX") == "true",
	}

	// Validate required configuration
	if config.APNSAuthKey == "" || config.APNSKeyID == "" ||
		config.APNSTeamID == "" || config.APNSBundleID == "" {
		return fmt.Errorf("missing required APNS configuration")
	}

	return nil
}

// APNSPayload represents the payload for an iOS push notification
type APNSPayload struct {
	Aps struct {
		Alert struct {
			Title    string `json:"title"`
			Subtitle string `json:"subtitle,omitempty"`
			Body     string `json:"body"`
		} `json:"alert"`
		Badge            int    `json:"badge,omitempty"`
		Sound            string `json:"sound,omitempty"`
		ContentAvailable int    `json:"content-available,omitempty"`
		MutableContent   int    `json:"mutable-content,omitempty"`
		Category         string `json:"category,omitempty"`
	} `json:"aps"`
	EventID uint `json:"event_id,omitempty"`
	TaskID  uint `json:"task_id,omitempty"`
}

// SendTaskCompletionNotification sends a push notification to the event organizer
// when a task is completed
func SendTaskCompletionNotification(deviceToken string, eventID uint, taskID uint, taskTitle string, completedBy string) error {
	// Create notification payload
	payload := APNSPayload{}
	payload.Aps.Alert.Title = "Task Completed"
	payload.Aps.Alert.Body = fmt.Sprintf("%s completed task: %s", completedBy, taskTitle)
	payload.Aps.Sound = "default"
	payload.Aps.Badge = 1
	payload.EventID = eventID
	payload.TaskID = taskID

	// Serialize payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal notification payload: %v", err)
	}

	// Determine which APNS endpoint to use
	apnsURL := APNSProductionURL
	if config.APNSIsSandbox {
		apnsURL = APNSSandboxURL
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/%s", apnsURL, deviceToken)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Get JWT token for authentication - in real implementation you'd use
	// the p8 file to generate a JWT token as described in Apple's documentation
	jwt, err := getAPNSAuthToken()
	if err != nil {
		return fmt.Errorf("failed to generate APNS JWT token: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("bearer %s", jwt))
	req.Header.Set("apns-topic", config.APNSBundleID)
	req.Header.Set("apns-push-type", "alert")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send push notification: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("APNS server returned non-200 status code: %d", resp.StatusCode)
	}

	return nil
}

// getAPNSAuthToken generates a JWT token for APNS authentication
func getAPNSAuthToken() (string, error) {
	// Read the p8 file containing the private key
	keyPath := config.APNSAuthKey
	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read p8 key file: %v", err)
	}

	// Parse the private key
	var privateKey *ecdsa.PrivateKey
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		// Try parsing the raw p8 file format
		parsedKey, err := x509.ParsePKCS8PrivateKey(keyBytes)
		if err != nil {
			return "", fmt.Errorf("failed to parse private key: %v", err)
		}

		var ok bool
		privateKey, ok = parsedKey.(*ecdsa.PrivateKey)
		if !ok {
			return "", fmt.Errorf("key is not an ECDSA private key")
		}
	} else {
		// If the key is in PEM format
		privateKey, err = jwt.ParseECPrivateKeyFromPEM(block.Bytes)
		if err != nil {
			return "", fmt.Errorf("failed to parse PEM private key: %v", err)
		}
	}

	// Create token with claims
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iss": config.APNSTeamID,                        // Team ID
		"iat": now.Unix(),                               // Issued at time
		"exp": now.Add(time.Hour).Unix(),                // Expiry time (1 hour from now)
		"kid": config.APNSKeyID,                         // Key ID
		"aud": "https://api.development.push.apple.com", // Audience
		"sub": config.APNSBundleID,                      // Subject (App Bundle ID)
	})

	// Sign the token with the private key
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT token: %v", err)
	}

	return signedToken, nil
}

// StoreDeviceToken stores a user's device token for push notifications
func StoreDeviceToken(userID uint, deviceToken string, deviceType string) error {
	// Check if the token already exists for this user and device type
	var existingToken models.DeviceToken
	result := db.Where("user_id = ? AND token = ?", userID, deviceToken).First(&existingToken)

	now := time.Now().Unix()

	if result.Error == nil {
		// Token exists, update last used time and ensure it's active
		existingToken.LastUsedAt = now
		existingToken.IsActive = true
		return db.Save(&existingToken).Error
	}

	// Token doesn't exist, create a new one
	newToken := models.DeviceToken{
		UserID:     userID,
		Token:      deviceToken,
		DeviceType: deviceType,
		IsActive:   true,
		CreatedAt:  now,
		LastUsedAt: now,
	}

	return db.Create(&newToken).Error
}

// NotifyEventOrganizer sends a push notification to the event organizer when a task is completed
func NotifyEventOrganizer(eventID uint, taskID uint, taskTitle string, completedByUserID uint) error {
	// Get the event to find the organizer
	var event models.Event
	if err := db.First(&event, eventID).Error; err != nil {
		return fmt.Errorf("failed to find event: %v", err)
	}

	// Get the user who completed the task for the notification text
	var completedByUser models.User
	if err := db.First(&completedByUser, completedByUserID).Error; err != nil {
		return fmt.Errorf("failed to find user: %v", err)
	}

	// Get the organizer's active device tokens
	var deviceTokens []models.DeviceToken
	if err := db.Where("user_id = ? AND is_active = ? AND device_type = ?",
		event.OrganizerID, true, "ios").Find(&deviceTokens).Error; err != nil {
		return fmt.Errorf("failed to find device tokens: %v", err)
	}

	// Send notification to all of the organizer's iOS devices
	for _, token := range deviceTokens {
		err := SendTaskCompletionNotification(
			token.Token,
			eventID,
			taskID,
			taskTitle,
			completedByUser.DisplayName,
		)
		if err != nil {
			// Log error but continue with other tokens
			fmt.Printf("Error sending notification to token %s: %v\n", token.Token, err)
		}
	}

	return nil
}
