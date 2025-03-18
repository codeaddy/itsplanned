package handlers

import (
	"itsplanned/models"
	"itsplanned/models/api"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func toAIChatResponse(chat *models.AIChat) *api.AIChatResponse {
	if chat == nil {
		return nil
	}
	return &api.AIChatResponse{
		ID:        chat.ID,
		UserID:    chat.UserID,
		CreatedAt: chat.CreatedAt,
	}
}

func toAIMessageResponse(message *models.AIMessage) *api.AIMessageResponse {
	if message == nil {
		return nil
	}
	return &api.AIMessageResponse{
		ID:        message.ID,
		ChatID:    message.ChatID,
		UserID:    message.UserID,
		Content:   message.Content,
		IsUser:    message.IsUser,
		CreatedAt: message.CreatedAt,
	}
}

// @Summary Start a new AI chat
// @Description Create a new AI chat session for the user
// @Tags ai-assistant
// @Produce json
// @Security BearerAuth
// @Success 200 {object} api.APIResponse{data=api.AIChatResponse} "Chat started successfully"
// @Failure 401 {object} api.APIResponse "Unauthorized"
// @Failure 500 {object} api.APIResponse "Failed to start chat"
// @Router /ai/chat [post]
func StartAIChat(c *gin.Context, db *gorm.DB) {
	userID, _ := c.Get("user_id")

	chat := models.AIChat{
		UserID: userID.(uint),
	}

	if err := db.Create(&chat).Error; err != nil {
		c.JSON(http.StatusInternalServerError, api.APIResponse{Error: "Failed to start chat"})
		return
	}

	c.JSON(http.StatusOK, api.APIResponse{
		Message: "Chat started",
		Data:    toAIChatResponse(&chat),
	})
}

// @Summary Send message to AI assistant
// @Description Send a message to the AI assistant and get a response
// @Tags ai-assistant
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body api.SendMessageRequest true "Message details"
// @Success 200 {object} api.SendMessageResponse "Message sent and response received"
// @Failure 400 {object} api.APIResponse "Invalid payload"
// @Failure 401 {object} api.APIResponse "Unauthorized"
// @Failure 500 {object} api.APIResponse "Failed to process message"
// @Router /ai/message [post]
func SendMessage(c *gin.Context, db *gorm.DB) {
	var request api.SendMessageRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, api.APIResponse{Error: "Invalid payload"})
		return
	}

	userID, _ := c.Get("user_id")

	message := models.AIMessage{
		ChatID:  request.ChatID,
		UserID:  userID.(uint),
		Content: request.Message,
		IsUser:  true,
	}

	if err := db.Create(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, api.APIResponse{Error: "Failed to save message"})
		return
	}

	// TODO: Integrate with actual AI service
	// For now, just return a mock response
	aiResponse := "Thank you for your message! I'm still learning how to help with event planning. Here are some theme suggestions based on your message..."

	aiMessage := models.AIMessage{
		ChatID:  request.ChatID,
		Content: aiResponse,
		IsUser:  false,
	}

	if err := db.Create(&aiMessage).Error; err != nil {
		c.JSON(http.StatusInternalServerError, api.APIResponse{Error: "Failed to save AI response"})
		return
	}

	c.JSON(http.StatusOK, api.SendMessageResponse{
		Message:  "Message sent",
		Response: *toAIMessageResponse(&aiMessage),
	})
}

// @Summary Get chat history
// @Description Get the message history for a specific chat
// @Tags ai-assistant
// @Produce json
// @Security BearerAuth
// @Param id path int true "Chat ID"
// @Success 200 {object} api.ChatHistoryResponse "Chat history retrieved successfully"
// @Failure 401 {object} api.APIResponse "Unauthorized"
// @Failure 403 {object} api.APIResponse "Forbidden - not your chat"
// @Failure 404 {object} api.APIResponse "Chat not found"
// @Failure 500 {object} api.APIResponse "Failed to fetch messages"
// @Router /ai/chat/{id} [get]
func GetChatHistory(c *gin.Context, db *gorm.DB) {
	chatID := c.Param("id")
	userID, _ := c.Get("user_id")

	var chat models.AIChat
	if err := db.First(&chat, chatID).Error; err != nil {
		c.JSON(http.StatusNotFound, api.APIResponse{Error: "Chat not found"})
		return
	}

	if chat.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, api.APIResponse{Error: "You can only access your own chats"})
		return
	}

	var messages []models.AIMessage
	if err := db.Where("chat_id = ?", chatID).Order("created_at asc").Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, api.APIResponse{Error: "Failed to fetch messages"})
		return
	}

	var response api.ChatHistoryResponse
	for _, msg := range messages {
		if msgResponse := toAIMessageResponse(&msg); msgResponse != nil {
			response.Messages = append(response.Messages, *msgResponse)
		}
	}

	c.JSON(http.StatusOK, response)
}
