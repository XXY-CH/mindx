package core

import "github.com/gin-gonic/gin"

// AuthProvider 认证插件接口
// 设计为可插拔的认证模块，默认使用 NoopAuthProvider（不需要认证）
// 用户可以通过实现此接口来添加自定义认证（如 JWT、API Key、OAuth 等）
type AuthProvider interface {
	// Name 返回认证提供者名称
	Name() string

	// Enabled 返回认证是否启用
	Enabled() bool

	// Middleware 返回 Gin 中间件，用于保护 API 路由
	// 返回 nil 表示不需要认证中间件
	Middleware() gin.HandlerFunc

	// PublicPaths 返回不需要认证的路径列表（如健康检查、登录等）
	PublicPaths() []string
}
