// Package auth 提供认证模块的依赖装配和路由自注册。
package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/nixwai/go-game-server/app/bootstrap"
	"github.com/nixwai/go-game-server/app/ginext"
	"github.com/nixwai/go-game-server/app/middleware"
	"github.com/nixwai/go-game-server/app/security"
)

// Register 构建认证模块依赖并注册路由。
func Register(rg *gin.RouterGroup, d *bootstrap.Deps) {
	h := Wire(d)
	RegisterRoutes(rg, h, d.Tokens)
}

// Wire 从共享依赖构建认证 Handler。
func Wire(d *bootstrap.Deps) *Handler {
	users := NewGormUserRepository(d.DB)
	hasher := security.PasswordHasher{
		Time:    d.Config.Argon2Time,
		Memory:  d.Config.Argon2Memory,
		Threads: d.Config.Argon2Threads,
		KeyLen:  d.Config.Argon2KeyLen,
		SaltLen: d.Config.Argon2SaltLen,
	}
	svc := NewService(users, hasher, d.Tokens)
	return NewHandler(svc, d.Crypto)
}

// RegisterRoutes 在指定路由组上注册认证路由。
func RegisterRoutes(rg *gin.RouterGroup, h *Handler, tokens *security.TokenManager) {
	g := rg.Group("/auth")
	g.GET("/public-key", ginext.Wrap(h.PublicKey))
	g.POST("/register", ginext.WrapJSON(h.Register))
	g.POST("/login", ginext.WrapJSON(h.Login))
	g.GET("/me", middleware.AuthRequired(tokens), ginext.Wrap(h.Me))
	g.POST("/password/update", middleware.AuthRequired(tokens), ginext.WrapJSON(h.ChangePassword))
	rg.GET("/admin/ping", middleware.AuthRequired(tokens), middleware.AdminOnly(), ginext.Wrap(h.AdminPing))
}
