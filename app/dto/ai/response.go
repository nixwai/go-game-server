package ai

import (
	"time"

	"github.com/nixwai/go-game-server/app/model"
)

// ModelResponse 是 AI 模型配置的对外响应结构。
type ModelResponse struct {
	// ID 是模型配置主键，默认模型使用 0。
	ID uint64 `json:"id"`
	// ModelName 是模型名称。
	ModelName string `json:"model_name"`
	// Status 是模型状态。
	Status string `json:"status"`
	// IsDefault 表示是否为配置文件提供的只读默认模型。
	IsDefault bool `json:"is_default"`
	// CreatedAt 是模型配置创建时间。
	CreatedAt time.Time `json:"created_at"`
}

// ProviderResponse 是 AI 产商配置的对外响应结构。
type ProviderResponse struct {
	// ID 是产商配置主键，默认产商使用 0。
	ID uint64 `json:"id"`
	// ProviderName 是产商名称。
	ProviderName string `json:"provider_name"`
	// BaseURL 是产商 API 的基础地址。
	BaseURL string `json:"base_url"`
	// Status 是产商状态。
	Status string `json:"status"`
	// IsDefault 表示是否为配置文件提供的只读默认产商。
	IsDefault bool `json:"is_default"`
	// HasAPIKey 表示该产商是否已配置 API Key。
	// 默认产商固定返回 true，用户产商根据数据库记录判断。
	HasAPIKey bool `json:"has_api_key"`
	// Models 是该产商下的模型列表。
	Models []ModelResponse `json:"models"`
	// CreatedAt 是产商配置创建时间。
	CreatedAt time.Time `json:"created_at"`
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
// hasAPIKey 表示该产商是否已配置 API Key，由 Service 层判断。
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
