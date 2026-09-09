package dto

import (
	"time"

	"github.com/nixwai/go-game-server/app/model"
)

// ModelResponse 是 AI 模型配置的对外响应结构。
type ModelResponse struct {
	ID        uint64    `json:"id"`
	ModelName string    `json:"model_name"`
	Status    string    `json:"status"`
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
}

// ProviderResponse 是 AI 产商配置的对外响应结构。
type ProviderResponse struct {
	ID           uint64          `json:"id"`
	ProviderName string          `json:"provider_name"`
	BaseURL      string          `json:"base_url"`
	Status       string          `json:"status"`
	IsDefault    bool            `json:"is_default"`
	HasAPIKey    bool            `json:"has_api_key"`
	Models       []ModelResponse `json:"models"`
	CreatedAt    time.Time       `json:"created_at"`
}

// NewModelResponse 将持久化模型转换为安全的接口响应对象。
func NewModelResponse(mdl model.AIModel, isDefault bool) ModelResponse {
	return ModelResponse{
		ID:        mdl.ID,
		ModelName: mdl.ModelName,
		Status:    mdl.Status,
		IsDefault: isDefault,
		CreatedAt: mdl.CreatedAt,
	}
}

// NewProviderResponse 将持久化产商和模型列表转换为安全的接口响应对象。
func NewProviderResponse(provider model.AIProvider, models []ModelResponse, isDefault bool, hasAPIKey bool) ProviderResponse {
	return ProviderResponse{
		ID:           provider.ID,
		ProviderName: provider.ProviderName,
		BaseURL:      provider.BaseURL,
		Status:       provider.Status,
		IsDefault:    isDefault,
		HasAPIKey:    hasAPIKey,
		Models:       models,
		CreatedAt:    provider.CreatedAt,
	}
}
