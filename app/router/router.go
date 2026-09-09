// Package router 集中注册服务的 HTTP 路由和中间件链。
package router

import (
	"github.com/gin-gonic/gin"
	"github.com/nixwai/go-game-server/app/bootstrap"
	"github.com/nixwai/go-game-server/app/ginext"
	"github.com/nixwai/go-game-server/app/middleware"
	"github.com/nixwai/go-game-server/app/module/ai"
	"github.com/nixwai/go-game-server/app/module/auth"
)

// health 返回服务存活状态，不需要身份认证。
func health(c *gin.Context) (gin.H, error) {
	return gin.H{"status": "ok"}, nil
}

// New 创建并配置 Gin 路由引擎。
func New(d *bootstrap.Deps) *gin.Engine {
	r := gin.New()
	// Recovery 防止未处理 panic 终止进程，RequestID 为日志和客户端提供关联标识。
	r.Use(gin.Recovery(), middleware.RequestID())
	r.GET("/health", ginext.Wrap(health))
	r.StaticFile("/docs/openapi.yaml", "docs/openapi.yaml")

	api := r.Group("/api/v1")
	auth.Register(api, d)
	ai.Register(api, d)
	return r
}
