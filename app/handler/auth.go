// Package handler 将 HTTP 请求转换为认证服务调用，并输出统一响应。
package handler

import (
	"github.com/gin-gonic/gin"
	auth "github.com/nixwai/go-game-server/app/dto/auth"
	"github.com/nixwai/go-game-server/app/middleware"
	"github.com/nixwai/go-game-server/app/service"
)

// AuthHandler 是认证相关接口的 HTTP 处理器。
type AuthHandler struct {
	// service 是认证业务服务。
	service *service.AuthService
}

// NewAuthHandler 创建认证 HTTP 处理器。
func NewAuthHandler(s *service.AuthService) *AuthHandler { return &AuthHandler{service: s} }

// Register 处理普通用户注册请求，返回新用户的脱敏信息。
func (h *AuthHandler) Register(c *gin.Context, req auth.RegisterRequest) (auth.UserResponse, error) {
	user, err := h.service.Register(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		return auth.UserResponse{}, err
	}
	return auth.NewUserResponse(user), nil
}

// Login 处理用户名密码登录请求，返回 JWT 和用户信息。
func (h *AuthHandler) Login(c *gin.Context, req auth.LoginRequest) (auth.LoginResponse, error) {
	result, err := h.service.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		return auth.LoginResponse{}, err
	}
	return auth.NewLoginResponse(result.Token, result.User), nil
}

// Me 返回当前登录用户的最新信息。
func (h *AuthHandler) Me(c *gin.Context) (auth.UserResponse, error) {
	userID := middleware.GetUserID(c)
	user, err := h.service.CurrentUser(c.Request.Context(), userID)
	if err != nil {
		return auth.UserResponse{}, err
	}
	return auth.NewUserResponse(user), nil
}

// AdminPing 是仅用于验证管理员权限中间件的示例接口。
func (h *AuthHandler) AdminPing(c *gin.Context) (gin.H, error) {
	return gin.H{"message": "管理员权限验证通过"}, nil
}
