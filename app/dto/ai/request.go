package ai

// CreateProviderRequest 是新增 AI 产商配置的请求。
type CreateProviderRequest struct {
	// ProviderName 是产商名称，由用户自由输入。
	ProviderName string `json:"provider_name" binding:"required"`
	// BaseURL 是产商 API 的基础地址。
	BaseURL string `json:"base_url" binding:"required"`
	// EncryptedAPIKey 是前端使用 RSA 公钥加密后的 base64 编码密文。
	EncryptedAPIKey string `json:"encrypted_api_key" binding:"required"`
}

// UpdateProviderRequest 是更新 AI 产商配置的请求，ID 必填，其余字段可选。
type UpdateProviderRequest struct {
	// ID 是目标产商 ID。
	ID uint64 `json:"id"`
	// ProviderName 是新的产商名称，不提供则保留原值。
	ProviderName *string `json:"provider_name,omitempty"`
	// BaseURL 是新的 API 基础地址，不提供则保留原值。
	BaseURL *string `json:"base_url,omitempty"`
	// EncryptedAPIKey 是新的 RSA 加密后的 API Key 密文，不提供则保留原值。
	EncryptedAPIKey *string `json:"encrypted_api_key,omitempty"`
	// Status 是新的状态值（active 或 disabled），不提供则保留原值。
	Status *string `json:"status,omitempty"`
}

// DeleteProviderRequest 是删除 AI 产商配置的请求。
type DeleteProviderRequest struct {
	// ID 是目标产商 ID。
	ID uint64 `json:"id"`
}

// CreateModelRequest 是新增 AI 模型配置的请求。
type CreateModelRequest struct {
	// ProviderID 是目标产商 ID。
	ProviderID uint64 `json:"provider_id" binding:"required"`
	// ModelName 是模型名称，由用户自由输入。
	ModelName string `json:"model_name" binding:"required"`
}

// UpdateModelRequest 是更新 AI 模型配置的请求，ID 必填，其余字段可选。
type UpdateModelRequest struct {
	// ID 是目标模型 ID。
	ID uint64 `json:"id"`
	// ModelName 是新的模型名称，不提供则保留原值。
	ModelName *string `json:"model_name,omitempty"`
	// Status 是新的状态值，不提供则保留原值。
	Status *string `json:"status,omitempty"`
}

// DeleteModelRequest 是删除 AI 模型配置的请求。
type DeleteModelRequest struct {
	// ID 是目标模型 ID。
	ID uint64 `json:"id"`
}
