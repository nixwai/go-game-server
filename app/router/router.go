// Package router 集中注册服务的 HTTP 路由和中间件链。
package router

import (
	"github.com/gin-gonic/gin"
	"github.com/nixwai/go-game-server/app/handler"
	"github.com/nixwai/go-game-server/app/middleware"
	"github.com/nixwai/go-game-server/app/security"
)

// New 创建并配置 Gin 路由引擎。
func New(authHandler *handler.AuthHandler, tokens *security.TokenManager) *gin.Engine {
	r := gin.New()
	// Recovery 防止未处理 panic 终止进程，RequestID 为日志和客户端提供关联标识。
	r.Use(gin.Recovery(), middleware.RequestID())
	r.GET("/health", handler.Health)
	r.StaticFile("/docs/openapi.yaml", "docs/openapi.yaml")

	api := r.Group("/api/v1")
	auth := api.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)
	auth.GET("/me", middleware.AuthRequired(tokens), authHandler.Me)
	api.GET("/admin/ping", middleware.AuthRequired(tokens), middleware.AdminOnly(), authHandler.AdminPing)
	return r
}
