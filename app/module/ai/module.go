// Package ai 提供 AI 模型管理模块的依赖装配和路由自注册。
package ai

import (
	"github.com/gin-gonic/gin"
	"github.com/nixwai/go-game-server/app/bootstrap"
	"github.com/nixwai/go-game-server/app/ginext"
	"github.com/nixwai/go-game-server/app/middleware"
	"github.com/nixwai/go-game-server/app/security"
)

// Register 构建 AI 模块依赖并注册路由。
func Register(rg *gin.RouterGroup, d *bootstrap.Deps) {
	h := Wire(d)
	RegisterRoutes(rg, h, d.Tokens)
}

// Wire 从共享依赖构建 AI 管理 Handler。
func Wire(d *bootstrap.Deps) *Handler {
	providers := NewGormProviderRepository(d.DB)
	models := NewGormModelRepository(d.DB)
	svc := NewService(providers, models, d.Crypto, d.Config.DefaultAI)
	return NewHandler(svc, d.Crypto)
}

// RegisterRoutes 在指定路由组上注册 AI 模型管理路由。
func RegisterRoutes(rg *gin.RouterGroup, h *Handler, tokens *security.TokenManager) {
	g := rg.Group("/ai", middleware.AuthRequired(tokens))
	g.GET("/public-key", ginext.Wrap(h.PublicKey))
	g.GET("/providers/list", ginext.Wrap(h.ListProviders))
	g.POST("/providers/create", ginext.WrapJSON(h.CreateProvider))
	g.POST("/providers/update", ginext.WrapJSON(h.UpdateProvider))
	g.POST("/providers/delete", ginext.WrapJSON(h.DeleteProvider))
	g.POST("/models/create", ginext.WrapJSON(h.CreateModel))
	g.POST("/models/update", ginext.WrapJSON(h.UpdateModel))
	g.POST("/models/delete", ginext.WrapJSON(h.DeleteModel))
}
