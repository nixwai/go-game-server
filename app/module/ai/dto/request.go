package dto

// CreateProviderRequest 是新增 AI 产商配置的请求。
type CreateProviderRequest struct {
	ProviderName    string `json:"provider_name" binding:"required"`
	BaseURL         string `json:"base_url" binding:"required"`
	EncryptedAPIKey string `json:"encrypted_api_key" binding:"required"`
}

// UpdateProviderRequest 是更新 AI 产商配置的请求，ID 必填，其余字段可选。
type UpdateProviderRequest struct {
	ID              uint64  `json:"id"`
	ProviderName    *string `json:"provider_name,omitempty"`
	BaseURL         *string `json:"base_url,omitempty"`
	EncryptedAPIKey *string `json:"encrypted_api_key,omitempty"`
	Status          *string `json:"status,omitempty"`
}

// DeleteProviderRequest 是删除 AI 产商配置的请求。
type DeleteProviderRequest struct {
	ID uint64 `json:"id"`
}

// CreateModelRequest 是新增 AI 模型配置的请求。
type CreateModelRequest struct {
	ProviderID uint64 `json:"provider_id" binding:"required"`
	ModelName  string `json:"model_name" binding:"required"`
}

// UpdateModelRequest 是更新 AI 模型配置的请求，ID 必填，其余字段可选。
type UpdateModelRequest struct {
	ID        uint64  `json:"id"`
	ModelName *string `json:"model_name,omitempty"`
	Status    *string `json:"status,omitempty"`
}

// DeleteModelRequest 是删除 AI 模型配置的请求。
type DeleteModelRequest struct {
	ID uint64 `json:"id"`
}
