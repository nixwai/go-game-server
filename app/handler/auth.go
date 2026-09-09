// Package handler 将 HTTP 请求转换为认证服务调用，并输出统一响应。
package handler

import (
	"github.com/gin-gonic/gin"
	auth "github.com/nixwai/go-game-server/app/dto/auth"
	"github.com/nixwai/go-game-server/app/middleware"
	"github.com/nixwai/go-game-server/app/response"
	"github.com/nixwai/go-game-server/app/service"
)

// AuthHandler 是认证相关接口的 HTTP 处理器。
type AuthHandler struct {
	// service 是认证业务服务。
	service *service.AuthService
}

// NewAuthHandler 创建认证 HTTP 处理器。
func NewAuthHandler(s *service.AuthService) *AuthHandler { return &AuthHandler{service: s} }

// Register 处理普通用户注册请求。
func (h *AuthHandler) Register(c *gin.Context, req auth.RegisterRequest) {
	user, err := h.service.Register(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Write(c, response.CodeOK, "成功", auth.NewUserResponse(user))
}

// Login 处理用户名密码登录请求，并返回 JWT。
func (h *AuthHandler) Login(c *gin.Context, req auth.LoginRequest) {
	result, err := h.service.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Write(c, response.CodeOK, "成功", auth.NewLoginResponse(result.Token, result.User))
}

// Me 返回当前登录用户的最新信息。
func (h *AuthHandler) Me(c *gin.Context) {
	userID := middleware.GetUserID(c)
	user, err := h.service.CurrentUser(c.Request.Context(), userID)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Write(c, response.CodeOK, "成功", auth.NewUserResponse(user))
}

// AdminPing 是仅用于验证管理员权限中间件的示例接口。
func (h *AuthHandler) AdminPing(c *gin.Context) {
	response.Write(c, response.CodeOK, "成功", gin.H{"message": "管理员权限验证通过"})
}
