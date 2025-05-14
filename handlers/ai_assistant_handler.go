package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"itsplanned/models/api"
	"itsplanned/services/yandex"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// @Summary Send message to Yandex GPT
// @Description Proxy request to Yandex GPT API and return the response
// @Tags ai-assistant
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body api.YandexGPTRequest true "Dialog history to send to Yandex GPT"
// @Success 200 {object} api.YandexGPTResponse "Response from Yandex GPT"
// @Failure 400 {object} api.APIResponse "Invalid payload"
// @Failure 401 {object} api.APIResponse "Unauthorized"
// @Failure 500 {object} api.APIResponse "Failed to process message"
// @Router /ai/message [post]
func SendToYandexGPT(c *gin.Context, db *gorm.DB) {
	var request api.YandexGPTRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Invalid payload"})
		return
	}

	folderID := os.Getenv("YANDEX_CATALOG_ID")

	// Get IAM token from token service
	iamToken, err := yandex.GetToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.APIResponse{Error: "Failed to get Yandex API token: " + err.Error()})
		return
	}

	if folderID == "" || iamToken == "" {
		c.JSON(http.StatusInternalServerError, api.APIResponse{Error: "Missing Yandex GPT API credentials"})
		return
	}

	yandexRequest := api.YandexGPTAPIRequest{
		ModelUri: fmt.Sprintf("gpt://%s/yandexgpt", folderID),
		CompletionOptions: struct {
			Stream           bool    `json:"stream"`
			Temperature      float64 `json:"temperature"`
			MaxTokens        string  `json:"maxTokens"`
			ReasoningOptions struct {
				Mode string `json:"mode"`
			} `json:"reasoningOptions"`
		}{
			Stream:      false,
			Temperature: 0.6,
			MaxTokens:   "2000",
			ReasoningOptions: struct {
				Mode string `json:"mode"`
			}{
				Mode: "DISABLED",
			},
		},
		Messages: request.Messages,
	}

	// Convert request to JSON
	requestBody, err := json.Marshal(yandexRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.APIResponse{Error: "Failed to prepare request"})
		return
	}

	// Send request to Yandex GPT API
	yandexGPTURL := "https://llm.api.cloud.yandex.net/foundationModels/v1/completion"
	req, err := http.NewRequest("POST", yandexGPTURL, bytes.NewBuffer(requestBody))
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.APIResponse{Error: "Failed to create request"})
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+iamToken)
	req.Header.Set("x-folder-id", folderID)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.APIResponse{Error: "Failed to send request to Yandex GPT"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, api.APIResponse{Error: fmt.Sprintf("Yandex GPT API returned error: %d", resp.StatusCode)})
		return
	}

	var yandexResponse api.YandexGPTAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&yandexResponse); err != nil {
		c.JSON(http.StatusInternalServerError, api.APIResponse{Error: "Failed to parse Yandex GPT response"})
		return
	}

	if len(yandexResponse.Result.Alternatives) == 0 {
		c.JSON(http.StatusInternalServerError, api.APIResponse{Error: "No response from Yandex GPT"})
		return
	}

	c.JSON(http.StatusOK, api.YandexGPTResponse{
		Message: yandexResponse.Result.Alternatives[0].Message.Text,
	})
}
