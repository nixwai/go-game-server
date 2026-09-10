// Package model 定义持久化模型和领域常量。
package model

import "time"

const (
	// RoleAdmin 表示拥有管理权限的用户角色。
	RoleAdmin = "admin"
	// RoleUser 表示普通用户角色。
	RoleUser = "user"
	// StatusActive 表示用户可以正常登录和访问服务。
	StatusActive = "active"
	// StatusDisabled 表示用户被禁用，不能登录。
	StatusDisabled = "disabled"
)

// User 是 user 表对应的持久化实体。
type User struct {
	// ID 是用户的自增主键。
	ID uint64 `gorm:"primaryKey;autoIncrement"`
	// Username 是用户登录名，必须唯一。
	Username string `gorm:"type:varchar(64);not null;uniqueIndex"`
	// PasswordHash 保存 Argon2id 哈希。
	PasswordHash string `gorm:"type:varchar(255);not null"`
	// Role 保存用户角色，例如 admin 或 user。
	Role string `gorm:"type:varchar(16);not null;index"`
	// Status 保存用户当前状态，例如 active 或 disabled。
	Status string `gorm:"type:varchar(16);not null;index"`
	// CreatedAt 是用户创建时间。
	CreatedAt time.Time
	// UpdatedAt 是用户最后更新时间。
	UpdatedAt time.Time
}

// TableName 返回 user 表名。
func (User) TableName() string { return "user" }
