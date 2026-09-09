// Package middleware 提供 JWT 鉴权和角色权限中间件。
package middleware

import (
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
			response.WriteError(c, response.NewError(response.CodeTokenInvalid, "需要认证", nil))
			c.Abort()
			return
		}
		claims, err := tokens.Parse(strings.TrimSpace(strings.TrimPrefix(header, prefix)))
		if err != nil {
			response.WriteError(c, response.NewError(response.CodeTokenInvalid, "令牌无效", nil))
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
		claims, ok := GetClaims(c)
		if !ok || claims.Role != model.RoleAdmin {
			response.WriteError(c, response.NewError(response.CodeForbidden, "权限不足", nil))
			c.Abort()
			return
		}
		c.Next()
	}
}

// GetClaims 从 Gin 上下文中读取 JWT Claims。
func GetClaims(c *gin.Context) (*security.Claims, bool) {
	claimsValue, exists := c.Get("claims")
	claims, ok := claimsValue.(*security.Claims)
	if !exists || !ok {
		return nil, false
	}
	return claims, true
}

// GetUserID 从 Gin 上下文中读取当前用户 ID。
func GetUserID(c *gin.Context) uint64 {
	claims, ok := GetClaims(c)
	if !ok {
		return 0
	}
	return claims.UserID
}
