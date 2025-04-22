package api

import "time"

// AIChatResponse represents an AI chat in API responses
type AIChatResponse struct {
	ID        uint      `json:"id" example:"1"`
	UserID    uint      `json:"user_id" example:"1"`
	CreatedAt time.Time `json:"created_at" example:"2024-03-16T12:00:00Z"`
}

// AIMessageResponse represents an AI message in API responses
type AIMessageResponse struct {
	ID        uint      `json:"id" example:"1"`
	ChatID    uint      `json:"chat_id" example:"1"`
	UserID    uint      `json:"user_id,omitempty" example:"1"`
	Content   string    `json:"content" example:"What theme would you suggest for a birthday party?"`
	IsUser    bool      `json:"is_user" example:"true"`
	CreatedAt time.Time `json:"created_at" example:"2024-03-16T12:00:00Z"`
}

// SendMessageRequest represents the request to send a message in an AI chat
type SendMessageRequest struct {
	ChatID  uint   `json:"chat_id" example:"1"`
	Message string `json:"message" example:"What theme would you suggest for a birthday party?"`
}

// SendMessageResponse represents the response after sending a message
type SendMessageResponse struct {
	Message  string            `json:"message" example:"Message sent"`
	Response AIMessageResponse `json:"response"`
}

// ChatHistoryResponse represents the response containing chat history
type ChatHistoryResponse struct {
	Messages []AIMessageResponse `json:"messages"`
}

// YandexGPTRequest represents the request to Yandex GPT API
type YandexGPTRequest struct {
	Messages []YandexGPTMessage `json:"messages"`
}

// YandexGPTResponse represents the response from Yandex GPT API
type YandexGPTResponse struct {
	Message string `json:"message" example:"For a birthday party, I would suggest several themes..."`
}

// YandexGPTAPIRequest represents the request body for Yandex GPT API
type YandexGPTAPIRequest struct {
	ModelUri          string `json:"modelUri"`
	CompletionOptions struct {
		Stream           bool    `json:"stream"`
		Temperature      float64 `json:"temperature"`
		MaxTokens        string  `json:"maxTokens"`
		ReasoningOptions struct {
			Mode string `json:"mode"`
		} `json:"reasoningOptions"`
	} `json:"completionOptions"`
	Messages []YandexGPTMessage `json:"messages"`
}

// YandexGPTMessage represents a message in the Yandex GPT conversation
type YandexGPTMessage struct {
	Role string `json:"role" example:"user"`
	Text string `json:"text" example:"What theme would you suggest for a birthday party?"`
}

// YandexGPTAPIResponse represents the response body from Yandex GPT API
type YandexGPTAPIResponse struct {
	Result struct {
		Alternatives []struct {
			Message struct {
				Role string `json:"role"`
				Text string `json:"text"`
			} `json:"message"`
			Status string `json:"status"`
		} `json:"alternatives"`
		Usage struct {
			InputTextTokens  string `json:"inputTextTokens"`
			CompletionTokens string `json:"completionTokens"`
			TotalTokens      string `json:"totalTokens"`
		} `json:"usage"`
		ModelVersion string `json:"modelVersion"`
	} `json:"result"`
}
