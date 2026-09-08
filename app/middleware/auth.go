// Package middleware 提供 JWT 鉴权和角色权限中间件。
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nixwai/go-game-server/app/model"
	"github.com/nixwai/go-game-server/app/response"
	"github.com/nixwai/go-game-server/app/security"
)

// AuthRequired 校验 Authorization Bearer JWT，并将 Claims 注入 Gin 上下文。
func AuthRequired(tokens *security.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		const prefix = "Bearer "
		if !strings.HasPrefix(header, prefix) {
			response.WriteError(c, response.NewError(http.StatusUnauthorized, response.CodeTokenInvalid, "authentication required", nil))
			c.Abort()
			return
		}
		// 去除 Bearer 前缀和多余空格后再解析，解析失败时不暴露底层 JWT 错误。
		claims, err := tokens.Parse(strings.TrimSpace(strings.TrimPrefix(header, prefix)))
		if err != nil {
			response.WriteError(c, response.NewError(http.StatusUnauthorized, response.CodeTokenInvalid, "invalid token", nil))
			c.Abort()
			return
		}
		c.Set("claims", claims)
		c.Next()
	}
}

// AdminOnly 要求当前请求已通过 JWT 鉴权且令牌角色为 admin。
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		claimsValue, exists := c.Get("claims")
		claims, ok := claimsValue.(*security.Claims)
		if !exists || !ok || claims.Role != model.RoleAdmin {
			response.WriteError(c, response.NewError(http.StatusForbidden, response.CodeForbidden, "forbidden", nil))
			c.Abort()
			return
		}
		c.Next()
	}
}
