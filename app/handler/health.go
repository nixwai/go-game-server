// Package handler 将 HTTP 请求转换为健康检查响应。
package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/nixwai/go-game-server/app/response"
)

// Health 返回服务存活状态，不需要身份认证。
func Health(c *gin.Context) {
	response.Write(c, response.CodeOK, "success", gin.H{"status": "ok"})
}
