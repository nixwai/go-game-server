// Package handler 将 HTTP 请求转换为认证服务调用，并输出统一响应。
package handler

import (
	"github.com/gin-gonic/gin"
	auth "github.com/nixwai/go-game-server/app/dto/auth"
	"github.com/nixwai/go-game-server/app/response"
	"github.com/nixwai/go-game-server/app/security"
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
func (h *AuthHandler) Register(c *gin.Context) {
	var req auth.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.WriteError(c, response.NewError(response.CodeValidation, "请求参数无效", err))
		return
	}
	user, err := h.service.Register(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Write(c, response.CodeOK, "成功", auth.NewUserResponse(user))
}

// Login 处理用户名密码登录请求，并返回 JWT。
func (h *AuthHandler) Login(c *gin.Context) {
	var req auth.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.WriteError(c, response.NewError(response.CodeValidation, "请求参数无效", err))
		return
	}
	result, err := h.service.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Write(c, response.CodeOK, "成功", auth.NewLoginResponse(result.Token, result.User))
}

// Me 返回当前登录用户的最新信息。
func (h *AuthHandler) Me(c *gin.Context) {
	claimsValue, exists := c.Get("claims")
	claims, ok := claimsValue.(*security.Claims)
	if !exists || !ok {
		response.WriteError(c, response.NewError(response.CodeTokenInvalid, "令牌无效", nil))
		return
	}
	user, err := h.service.CurrentUser(c.Request.Context(), claims.UserID)
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
