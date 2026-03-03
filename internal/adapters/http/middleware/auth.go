package middleware

import (
	"mindx/internal/core"
	"strings"

	"github.com/gin-gonic/gin"
)

// Auth 认证中间件
// 使用 AuthProvider 接口实现插件化认证
// 当 AuthProvider 未启用时，所有请求直接放行
func Auth(provider core.AuthProvider) gin.HandlerFunc {
	if provider == nil || !provider.Enabled() {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	authMiddleware := provider.Middleware()
	if authMiddleware == nil {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	publicPaths := provider.PublicPaths()

	return func(c *gin.Context) {
		// 检查是否为公开路径（不需要认证）
		reqPath := c.Request.URL.Path
		for _, pp := range publicPaths {
			if strings.HasPrefix(reqPath, pp) {
				c.Next()
				return
			}
		}

		// 执行认证
		authMiddleware(c)
	}
}
