package model

import "time"

const (
	// DefaultAIProviderID 是只读默认 AI 产商的虚拟 ID，不对应数据库记录。
	DefaultAIProviderID = 0
	// DefaultAIModelID 是默认 AI 模型的虚拟 ID，不对应数据库记录。
	DefaultAIModelID = 0
)

// AIProvider 是 ai_provider 表对应的持久化实体。
type AIProvider struct {
	// ID 是产商配置的自增主键。
	ID uint64 `gorm:"primaryKey;autoIncrement"`
	// UserID 是产商配置所属用户的主键。
	UserID uint64 `gorm:"not null;index"`
	// ProviderName 是产商名称，由用户自由输入。
	ProviderName string `gorm:"type:varchar(128);not null"`
	// BaseURL 是该产商 API 的基础地址。
	BaseURL string `gorm:"type:varchar(512);not null"`
	// APIKeyEncrypted 保存 AES-256-GCM 加密后的 API Key 密文。
	APIKeyEncrypted string `gorm:"type:varchar(1024);not null"`
	// Status 保存产商当前状态，active 或 disabled。
	Status string `gorm:"type:varchar(16);not null;index"`
	// CreatedAt 是产商配置创建时间。
	CreatedAt time.Time
	// UpdatedAt 是产商配置最后更新时间。
	UpdatedAt time.Time
}

// TableName 返回 ai_provider 表名。
func (AIProvider) TableName() string { return "ai_provider" }
