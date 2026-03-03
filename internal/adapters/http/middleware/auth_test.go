package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"mindx/internal/usecase/auth"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuth_NilProvider(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Auth(nil))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
}

func TestAuth_NoopProvider(t *testing.T) {
	gin.SetMode(gin.TestMode)
	noop := auth.NewNoopProvider()

	router := gin.New()
	router.Use(Auth(noop))
	router.GET("/api/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
}

// mockAuthProvider simulates a custom auth plugin for testing
type mockAuthProvider struct {
	enabled bool
}

func (m *mockAuthProvider) Name() string { return "mock" }
func (m *mockAuthProvider) Enabled() bool { return m.enabled }
func (m *mockAuthProvider) PublicPaths() []string {
	return []string{"/api/health", "/api/auth/"}
}
func (m *mockAuthProvider) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token != "Bearer test-token" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}

func TestAuth_EnabledProvider_ProtectedRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &mockAuthProvider{enabled: true}

	router := gin.New()
	router.Use(Auth(provider))
	router.GET("/api/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// Without token - should be rejected
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/test", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// With valid token - should pass
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuth_EnabledProvider_PublicRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &mockAuthProvider{enabled: true}

	router := gin.New()
	router.Use(Auth(provider))
	router.GET("/api/health", func(c *gin.Context) {
		c.String(http.StatusOK, "healthy")
	})

	// Public routes should pass without token
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/health", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "healthy", w.Body.String())
}

func TestAuth_DisabledProvider(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &mockAuthProvider{enabled: false}

	router := gin.New()
	router.Use(Auth(provider))
	router.GET("/api/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// Disabled provider should pass all requests
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/test", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
