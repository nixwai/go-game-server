// Package ginext 提供 Gin 框架扩展工具。
package ginext

import (
	"github.com/gin-gonic/gin"
	"github.com/nixwai/go-game-server/app/response"
)

// BindJSON 将 JSON Body 绑定到请求结构体后调用处理函数，绑定失败时返回参数校验错误。
func BindJSON[T any](h func(*gin.Context, T)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req T
		if err := c.ShouldBindJSON(&req); err != nil {
			response.WriteError(c, response.NewError(response.CodeValidation, "请求参数无效", err))
			return
		}
		h(c, req)
	}
}
