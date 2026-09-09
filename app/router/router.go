// Package router 集中注册服务的 HTTP 路由和中间件链。
package router

import (
	"github.com/gin-gonic/gin"
	"github.com/nixwai/go-game-server/app/ginext"
	"github.com/nixwai/go-game-server/app/handler"
	"github.com/nixwai/go-game-server/app/middleware"
	"github.com/nixwai/go-game-server/app/security"
)

// New 创建并配置 Gin 路由引擎。
func New(authHandler *handler.AuthHandler, aiHandler *handler.AIHandler, tokens *security.TokenManager) *gin.Engine {
	r := gin.New()
	// Recovery 防止未处理 panic 终止进程，RequestID 为日志和客户端提供关联标识。
	r.Use(gin.Recovery(), middleware.RequestID())
	r.GET("/health", ginext.Wrap(handler.Health))
	r.StaticFile("/docs/openapi.yaml", "docs/openapi.yaml")

	api := r.Group("/api/v1")
	auth := api.Group("/auth")
	auth.POST("/register", ginext.WrapJSON(authHandler.Register))
	auth.POST("/login", ginext.WrapJSON(authHandler.Login))
	auth.GET("/me", middleware.AuthRequired(tokens), ginext.Wrap(authHandler.Me))
	api.GET("/admin/ping", middleware.AuthRequired(tokens), middleware.AdminOnly(), ginext.Wrap(authHandler.AdminPing))

	// AI 模型管理路由组，全部需要 JWT 认证。
	aiGroup := api.Group("/ai", middleware.AuthRequired(tokens))
	aiGroup.GET("/public-key", ginext.Wrap(aiHandler.PublicKey))
	aiGroup.GET("/providers/list", ginext.Wrap(aiHandler.ListProviders))
	aiGroup.POST("/providers/create", ginext.WrapJSON(aiHandler.CreateProvider))
	aiGroup.POST("/providers/update", ginext.WrapJSON(aiHandler.UpdateProvider))
	aiGroup.POST("/providers/delete", ginext.WrapJSON(aiHandler.DeleteProvider))
	aiGroup.POST("/models/create", ginext.WrapJSON(aiHandler.CreateModel))
	aiGroup.POST("/models/update", ginext.WrapJSON(aiHandler.UpdateModel))
	aiGroup.POST("/models/delete", ginext.WrapJSON(aiHandler.DeleteModel))

	return r
}
