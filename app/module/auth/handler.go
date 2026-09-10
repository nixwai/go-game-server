// Package auth 提供用户注册、登录和密码修改等认证接口。
package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/nixwai/go-game-server/app/middleware"
	"github.com/nixwai/go-game-server/app/module/auth/dto"
	"github.com/nixwai/go-game-server/app/response"
)

// TransportDecrypter 解密前端使用 RSA 公钥加密的传输层密文。
type TransportDecrypter interface {
	// PublicKeyPEM 返回 RSA 公钥的 PEM 字符串，供前端加密使用。
	PublicKeyPEM() (string, error)
	// DecryptTransport 解密 base64 编码的 RSA-OAEP 密文，返回明文。
	DecryptTransport(b64Ciphertext string) (string, error)
}

// Handler 是认证相关接口的 HTTP 处理器。
type Handler struct {
	service   *Service
	decrypter TransportDecrypter
}

// NewHandler 创建认证 HTTP 处理器。
func NewHandler(s *Service, decrypter TransportDecrypter) *Handler {
	return &Handler{service: s, decrypter: decrypter}
}

// PublicKey 返回 RSA 公钥的 PEM 字符串，供前端加密密码和 API Key 使用。
func (h *Handler) PublicKey(c *gin.Context) (gin.H, error) {
	pem, err := h.decrypter.PublicKeyPEM()
	if err != nil {
		return nil, response.NewError(response.CodeInternal, "服务器内部错误", err)
	}
	return gin.H{"public_key": pem}, nil
}

// Register 处理普通用户注册请求，解密密码后调用 Service 完成注册。
func (h *Handler) Register(c *gin.Context, req dto.RegisterRequest) (dto.UserResponse, error) {
	password, err := h.decrypter.DecryptTransport(req.Password)
	if err != nil {
		return dto.UserResponse{}, response.NewError(response.CodeDecryptFailed, "密码解密失败", err)
	}
	user, err := h.service.Register(c.Request.Context(), req.Username, password)
	if err != nil {
		return dto.UserResponse{}, err
	}
	return dto.NewUserResponse(user), nil
}

// Login 处理用户名密码登录请求，解密密码后调用 Service 完成认证。
func (h *Handler) Login(c *gin.Context, req dto.LoginRequest) (dto.LoginResponse, error) {
	password, err := h.decrypter.DecryptTransport(req.Password)
	if err != nil {
		return dto.LoginResponse{}, response.NewError(response.CodeDecryptFailed, "密码解密失败", err)
	}
	result, err := h.service.Login(c.Request.Context(), req.Username, password)
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

// ChangePassword 处理当前用户修改密码请求，解密新旧密码后调用 Service。
func (h *Handler) ChangePassword(c *gin.Context, req dto.ChangePasswordRequest) (gin.H, error) {
	userID := middleware.GetUserID(c)
	oldPassword, err := h.decrypter.DecryptTransport(req.OldPassword)
	if err != nil {
		return nil, response.NewError(response.CodeDecryptFailed, "密码解密失败", err)
	}
	newPassword, err := h.decrypter.DecryptTransport(req.NewPassword)
	if err != nil {
		return nil, response.NewError(response.CodeDecryptFailed, "密码解密失败", err)
	}
	if err := h.service.ChangePassword(c.Request.Context(), userID, oldPassword, newPassword); err != nil {
		return nil, err
	}
	return gin.H{}, nil
}

// AdminPing 是仅用于验证管理员权限中间件的示例接口。
func (h *Handler) AdminPing(c *gin.Context) (gin.H, error) {
	return gin.H{"message": "管理员权限验证通过"}, nil
}
