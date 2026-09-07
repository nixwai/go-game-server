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

// User 是 users 表对应的持久化实体。
type User struct {
	// ID 是用户的自增主键。
	ID uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	// Username 是用户登录名，必须唯一。
	Username string `gorm:"type:varchar(64);not null;uniqueIndex" json:"username"`
	// PasswordHash 保存 Argon2id 哈希，禁止序列化到 JSON。
	PasswordHash string `gorm:"type:varchar(255);not null" json:"-"`
	// Role 保存用户角色，例如 admin 或 user。
	Role string `gorm:"type:varchar(16);not null;index" json:"role"`
	// Status 保存用户当前状态，例如 active 或 disabled。
	Status string `gorm:"type:varchar(16);not null;index" json:"status"`
	// CreatedAt 是用户创建时间。
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt 是用户最后更新时间。
	UpdatedAt time.Time `json:"updated_at"`
}

// PublicUser 是可以安全返回给客户端的用户信息，不包含密码哈希。
type PublicUser struct {
	ID        uint64    `json:"id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// Public 将持久化用户转换为脱敏后的公开用户对象。
func (u User) Public() PublicUser {
	return PublicUser{ID: u.ID, Username: u.Username, Role: u.Role, Status: u.Status, CreatedAt: u.CreatedAt}
}
