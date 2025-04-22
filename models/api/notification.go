package api

// RegisterDeviceTokenRequest represents the request to register a device token
type RegisterDeviceTokenRequest struct {
	DeviceToken string `json:"device_token" binding:"required" example:"6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b"`
	DeviceType  string `json:"device_type" binding:"required" example:"ios"` // ios, android, web
}
