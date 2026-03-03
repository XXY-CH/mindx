package auth

import "github.com/gin-gonic/gin"

// NoopProvider 默认认证提供者（不启用认证）
// 契合 MindX「轻量化仿生大脑」的设计理念，让不需要认证的用户保持部署的简洁
type NoopProvider struct{}

// NewNoopProvider 创建默认的无认证提供者
func NewNoopProvider() *NoopProvider {
	return &NoopProvider{}
}

func (p *NoopProvider) Name() string {
	return "noop"
}

func (p *NoopProvider) Enabled() bool {
	return false
}

func (p *NoopProvider) Middleware() gin.HandlerFunc {
	return nil
}

func (p *NoopProvider) PublicPaths() []string {
	return nil
}
