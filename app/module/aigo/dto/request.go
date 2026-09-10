package dto

// GoSign 表示棋盘上的棋子标记：1 为黑子，-1 为白子，0 为空位。
type GoSign int8

// GoLayout 是围棋棋盘的二维布局数据。
type GoLayout [][]GoSign

// GoVertex 是围棋棋盘的坐标，格式为 [row, col]。
type GoVertex [2]int

// PlayerSign 是可落子的棋子标记，不包含空位标记 0。
type PlayerSign GoSign

// KoInfo 描述劫子信息。
type KoInfo struct {
	// Sign 是劫子发生的棋子标记。
	Sign GoSign `json:"sign"`
	// Vertex 是劫子发生时的顶点。
	Vertex GoVertex `json:"vertex"`
}

// UpdateGameSettingRequest 是更新对弈配置的请求，所有字段可选。
type UpdateGameSettingRequest struct {
	// ActiveModelID 是用户激活的 AI 模型 ID，0 表示默认模型。
	ActiveModelID *uint64 `json:"active_model_id,omitempty"`
	// AllowAIEndGame 控制是否允许 AI 申请结束游戏。
	AllowAIEndGame *bool `json:"allow_ai_end_game,omitempty"`
}

// AnalyzeRequest 是分析棋局的请求，直接包含棋局快照字段。
type AnalyzeRequest struct {
	// Size 是棋盘尺寸，可为数字或字符串。
	Size any `json:"size" binding:"required"`
	// Layout 是棋盘布局。
	Layout GoLayout `json:"layout" binding:"required"`
	// Player 是当前执棋方。
	Player PlayerSign `json:"player" binding:"required"`
	// Ko 是劫子信息，无劫时省略。
	Ko *KoInfo `json:"ko,omitempty"`
	// LatestVertex 是最新一手棋子的棋盘坐标。
	LatestVertex *GoVertex `json:"latestVertex,omitempty"`
}
