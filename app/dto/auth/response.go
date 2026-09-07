package auth

import (
	"time"

	"github.com/nixwai/go-game-server/app/model"
)

// UserResponse 是认证接口对外返回的脱敏用户信息。
type UserResponse struct {
	// ID 是用户主键。
	ID uint64 `json:"id"`
	// Username 是用户登录名。
	Username string `json:"username"`
	// Role 是用户角色。
	Role string `json:"role"`
	// Status 是用户状态。
	Status string `json:"status"`
	// CreatedAt 是用户创建时间。
	CreatedAt time.Time `json:"created_at"`
}

// LoginResponse 是登录成功后的接口响应数据。
type LoginResponse struct {
	// Token 是客户端后续请求使用的 Bearer Access Token。
	Token string `json:"token"`
	// User 是登录用户的脱敏信息。
	User UserResponse `json:"user"`
}

// NewUserResponse 将持久化用户转换为安全的接口响应对象。
func NewUserResponse(user model.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Role:      user.Role,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
	}
}

// NewLoginResponse 将认证服务结果转换为登录接口响应对象。
func NewLoginResponse(token string, user model.User) LoginResponse {
	return LoginResponse{Token: token, User: NewUserResponse(user)}
}
