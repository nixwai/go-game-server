package model

import "time"

// AIModel 是 ai_model 表对应的持久化实体。
type AIModel struct {
	// ID 是模型配置的自增主键。
	ID uint64 `gorm:"primaryKey;autoIncrement"`
	// ProviderID 是该模型所属产商的主键。
	ProviderID uint64 `gorm:"not null;index"`
	// ModelName 是模型名称，由用户自由输入。
	ModelName string `gorm:"type:varchar(128);not null"`
	// Status 保存模型当前状态，active 或 disabled。
	Status string `gorm:"type:varchar(16);not null;index"`
	// CreatedAt 是模型配置创建时间。
	CreatedAt time.Time
	// UpdatedAt 是模型配置最后更新时间。
	UpdatedAt time.Time
}

// TableName 返回 ai_model 表名。
func (AIModel) TableName() string { return "ai_model" }
