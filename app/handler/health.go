// Package handler 将 HTTP 请求转换为健康检查响应。
package handler

import "github.com/gin-gonic/gin"

// Health 返回服务存活状态，不需要身份认证。
func Health(c *gin.Context) (gin.H, error) {
	return gin.H{"status": "ok"}, nil
}
