package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/nixwai/go-game-server/app/middleware"
	"github.com/nixwai/go-game-server/app/module/auth/dto"
)

// Handler 是认证相关接口的 HTTP 处理器。
type Handler struct {
	service *Service
}

// NewHandler 创建认证 HTTP 处理器。
func NewHandler(s *Service) *Handler { return &Handler{service: s} }

// Register 处理普通用户注册请求，返回新用户的脱敏信息。
func (h *Handler) Register(c *gin.Context, req dto.RegisterRequest) (dto.UserResponse, error) {
	user, err := h.service.Register(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		return dto.UserResponse{}, err
	}
	return dto.NewUserResponse(user), nil
}

// Login 处理用户名密码登录请求，返回 JWT 和用户信息。
func (h *Handler) Login(c *gin.Context, req dto.LoginRequest) (dto.LoginResponse, error) {
	result, err := h.service.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		return dto.LoginResponse{}, err
	}
	return dto.NewLoginResponse(result.Token, result.User), nil
}

// Me 返回当前登录用户的最新信息。
func (h *Handler) Me(c *gin.Context) (dto.UserResponse, error) {
	userID := middleware.GetUserID(c)
	user, err := h.service.CurrentUser(c.Request.Context(), userID)
	if err != nil {
		return dto.UserResponse{}, err
	}
	return dto.NewUserResponse(user), nil
}

// AdminPing 是仅用于验证管理员权限中间件的示例接口。
func (h *Handler) AdminPing(c *gin.Context) (gin.H, error) {
	return gin.H{"message": "管理员权限验证通过"}, nil
}
