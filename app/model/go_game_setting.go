package model

import "time"

// GoGameSetting 是 go_game_setting 表对应的持久化实体。
type GoGameSetting struct {
	// ID 是对弈配置的自增主键。
	ID uint64 `gorm:"primaryKey;autoIncrement"`
	// UserID 是配置所属用户的主键，每个用户仅有一条记录。
	UserID uint64 `gorm:"not null;uniqueIndex"`
	// ActiveModelID 是用户激活的 AI 模型 ID，0 表示使用默认模型。
	ActiveModelID uint64 `gorm:"not null;default:0"`
	// AllowAIEndGame 控制是否允许 AI 返回结束申请而非落子。
	AllowAIEndGame bool `gorm:"not null;default:true"`
	// CreatedAt 是配置创建时间。
	CreatedAt time.Time
	// UpdatedAt 是配置最后更新时间。
	UpdatedAt time.Time
}

// TableName 返回 go_game_setting 表名。
func (GoGameSetting) TableName() string { return "go_game_setting" }
