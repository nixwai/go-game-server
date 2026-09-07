// Package handler 将 HTTP 请求转换为认证服务调用，并输出统一响应。
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nixwai/go-game-server/app/model"
	"github.com/nixwai/go-game-server/app/response"
	"github.com/nixwai/go-game-server/app/security"
	"github.com/nixwai/go-game-server/app/service"
)

// AuthHandler 是认证相关接口的 HTTP 处理器。
type AuthHandler struct {
	// service 是认证业务服务。
	service *service.AuthService
}

// authRequest 是注册和登录共用的请求体。
type authRequest struct {
	// Username 是用户登录名。
	Username string `json:"username"`
	// Password 是用户明文密码，仅在请求处理期间使用。
	Password string `json:"password"`
	// Role 仅用于识别并拒绝客户端尝试创建管理员。
	Role string `json:"role"`
}

// NewAuthHandler 创建认证 HTTP 处理器。
func NewAuthHandler(s *service.AuthService) *AuthHandler { return &AuthHandler{service: s} }

// Register 处理普通用户注册请求。
func (h *AuthHandler) Register(c *gin.Context) {
	var req authRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.WriteError(c, response.NewError(400, response.CodeValidation, "invalid request body", err))
		return
	}
	// 注册接口永远由 Service 创建 user 角色；显式提交其他角色直接拒绝。
	if req.Role != "" && req.Role != model.RoleUser {
		response.WriteError(c, response.NewError(400, response.CodeValidation, "only user role can be registered", nil))
		return
	}
	user, err := h.service.Register(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Write(c, http.StatusCreated, response.CodeOK, "success", user)
}

// Login 处理用户名密码登录请求，并返回 JWT。
func (h *AuthHandler) Login(c *gin.Context) {
	var req authRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.WriteError(c, response.NewError(400, response.CodeValidation, "invalid request body", err))
		return
	}
	result, err := h.service.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Write(c, http.StatusOK, response.CodeOK, "success", result)
}

// Me 返回当前登录用户的最新信息。
func (h *AuthHandler) Me(c *gin.Context) {
	claimsValue, exists := c.Get("claims")
	claims, ok := claimsValue.(*security.Claims)
	if !exists || !ok {
		response.WriteError(c, response.NewError(401, response.CodeTokenInvalid, "invalid token", nil))
		return
	}
	user, err := h.service.CurrentUser(c.Request.Context(), claims.UserID)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Write(c, http.StatusOK, response.CodeOK, "success", user)
}

// AdminPing 是仅用于验证管理员权限中间件的示例接口。
func (h *AuthHandler) AdminPing(c *gin.Context) {
	response.Write(c, http.StatusOK, response.CodeOK, "success", gin.H{"message": "admin access granted"})
}

// Health 返回服务存活状态，不需要身份认证。
func Health(c *gin.Context) {
	response.Write(c, http.StatusOK, response.CodeOK, "success", gin.H{"status": "ok"})
}
