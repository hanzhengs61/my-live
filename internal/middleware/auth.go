package middleware

import (
	"context"
	"my-live/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware JWT 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 header Authorization 获取 token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
			return
		}

		// 验证 token
		claims, err := utils.ValidateJWT(authHeader)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token 无效或已过期"})
			return
		}

		// 把 user_id 存到 gin.Context，供后续 handler 使用
		userID, _ := (*claims)["user_id"].(float64)
		username, _ := (*claims)["username"].(string)
		c.Set("user_id", uint(userID))
		c.Set("username", username)
		c.Request = c.Request.WithContext(
			context.WithValue(c.Request.Context(), "gin_context", c),
		)
		// 通过，继续执行后续 handler
		c.Next()
	}
}
