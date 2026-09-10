// Package aigo 提供围棋对弈功能模块的依赖装配和路由自注册。
package aigo

import (
	"github.com/gin-gonic/gin"
	"github.com/nixwai/go-game-server/app/bootstrap"
	"github.com/nixwai/go-game-server/app/ginext"
	"github.com/nixwai/go-game-server/app/middleware"
	"github.com/nixwai/go-game-server/app/module/ai"
	"github.com/nixwai/go-game-server/app/security"
)

// Register 构建围棋对弈模块依赖并注册路由。
func Register(rg *gin.RouterGroup, d *bootstrap.Deps) {
	h := Wire(d)
	RegisterRoutes(rg, h, d.Tokens)
}

// Wire 从共享依赖构建围棋对弈 Handler。
func Wire(d *bootstrap.Deps) *Handler {
	settings := NewGormGameSettingRepository(d.DB)
	providers := ai.NewGormProviderRepository(d.DB)
	models := ai.NewGormModelRepository(d.DB)
	client := NewHTTPLLMClient()
	svc := NewService(settings, providers, models, d.Crypto, client, d.Config.DefaultAI, d.Config.GoLLMTimeout)
	return NewHandler(svc)
}

// RegisterRoutes 在指定路由组上注册围棋对弈路由。
func RegisterRoutes(rg *gin.RouterGroup, h *Handler, tokens *security.TokenManager) {
	g := rg.Group("/ai/go", middleware.AuthRequired(tokens))
	g.GET("/setting", ginext.Wrap(h.GetSetting))
	g.POST("/setting/update", ginext.WrapJSON(h.UpdateSetting))
	g.POST("/analyze", ginext.WrapJSON(h.Analyze))
}
