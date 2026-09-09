// Package ginext 提供 Gin 框架扩展工具。
package ginext

import (
	"github.com/gin-gonic/gin"
	"github.com/nixwai/go-game-server/app/response"
)

// writeResult 按返回的错误是否为空输出失败或成功响应。
func writeResult[O any](c *gin.Context, out O, err error) {
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Write(c, response.CodeOK, "成功", out)
}

// Wrap 包装无请求参数的处理函数，统一输出成功或失败响应。
func Wrap[O any](fn func(*gin.Context) (O, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		out, err := fn(c)
		writeResult(c, out, err)
	}
}

// WrapJSON 绑定 JSON 请求体后包装处理函数，统一输出成功或失败响应。
func WrapJSON[Req any, O any](fn func(*gin.Context, Req) (O, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Req
		if err := c.ShouldBindJSON(&req); err != nil {
			response.WriteError(c, response.NewError(response.CodeValidation, "请求参数无效", err))
			return
		}
		out, err := fn(c, req)
		writeResult(c, out, err)
	}
}
