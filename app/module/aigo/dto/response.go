package dto

// GameSettingResponse 是对弈配置的对外响应结构。
type GameSettingResponse struct {
	// ActiveModelID 是用户激活的 AI 模型 ID。
	ActiveModelID uint64 `json:"active_model_id"`
	// AllowAIEndGame 表示是否允许 AI 申请结束游戏。
	AllowAIEndGame bool `json:"allow_ai_end_game"`
	// ModelName 是激活模型的名称。
	ModelName string `json:"model_name"`
	// IsDefault 表示激活模型是否为只读默认模型。
	IsDefault bool `json:"is_default"`
	// HasAPIKey 表示该模型关联的产商是否已配置 API Key。
	HasAPIKey bool `json:"has_api_key"`
}

// AnalyzeResponse 是棋局分析的响应结构。
type AnalyzeResponse struct {
	// Action 是 AI 决定的动作：move 或 end_game。
	Action string `json:"action"`
	// Vertex 是落子坐标，仅当 Action 为 move 时存在。
	Vertex *GoVertex `json:"vertex,omitempty"`
}
