package middleware

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// AuthConfig 认证配置
type AuthConfig struct {
	// SecretKey 管理员密钥(可以是bcrypt哈希或明文)
	SecretKey string

	// AllowRemote 是否允许远程访问
	AllowRemote bool

	// LocalPassword 本地访问可选密码
	LocalPassword string
}

// AdminAuth 管理员认证中间件
// 借鉴CLIProxyAPI的设计,支持:
// - Bearer token 和 X-Admin-API-Key 两种认证方式
// - 环境变量和配置文件的secret
// - 本地客户端可选密码
// - 失败尝试限制和IP封禁
type AdminAuth struct {
	config         *AuthConfig
	envSecret      string
	mu             sync.Mutex
	failedAttempts map[string]*attemptInfo
}

type attemptInfo struct {
	count        int
	blockedUntil time.Time
	lastActivity time.Time
}

// NewAdminAuth 创建管理员认证中间件
func NewAdminAuth(secretKey string, allowRemote bool) *AdminAuth {
	envSecret := strings.TrimSpace(os.Getenv("ADMIN_API_KEY"))

	return &AdminAuth{
		config: &AuthConfig{
			SecretKey:   secretKey,
			AllowRemote: allowRemote,
		},
		envSecret:      envSecret,
		failedAttempts: make(map[string]*attemptInfo),
	}
}

// Middleware 返回Gin中间件函数
func (a *AdminAuth) Middleware() gin.HandlerFunc {
	const maxFailures = 5
	const banDuration = 30 * time.Minute

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		localClient := clientIP == "127.0.0.1" || clientIP == "::1"

		// 检查IP封禁
		if !localClient {
			a.mu.Lock()
			ai := a.failedAttempts[clientIP]
			if ai != nil && !ai.blockedUntil.IsZero() {
				if time.Now().Before(ai.blockedUntil) {
					a.mu.Unlock()
					remaining := time.Until(ai.blockedUntil).Round(time.Second)
					c.JSON(http.StatusForbidden, gin.H{
						"code":    403,
						"message": fmt.Sprintf("IP已被封禁,请在%s后重试", remaining),
						"error":   "ip_banned",
					})
					c.Abort()
					return
				}
				// 封禁过期,重置状态
				ai.blockedUntil = time.Time{}
				ai.count = 0
			}
			a.mu.Unlock()
		}

		// 检查远程访问权限
		if !localClient && !a.config.AllowRemote && a.envSecret == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "远程管理已禁用",
				"error":   "remote_disabled",
			})
			c.Abort()
			return
		}

		// 检查是否配置了密钥
		secretKey := a.config.SecretKey
		if secretKey == "" && a.envSecret == "" {
			// 未配置密钥,跳过认证(开发环境)
			c.Next()
			return
		}

		// 获取认证token
		token := a.extractToken(c)
		if token == "" {
			if !localClient {
				a.recordFailure(clientIP)
			}
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "缺少管理员API密钥",
				"error":   "missing_token",
			})
			c.Abort()
			return
		}

		// 验证token
		valid := false

		// 1. 检查环境变量secret
		if a.envSecret != "" && subtle.ConstantTimeCompare([]byte(token), []byte(a.envSecret)) == 1 {
			valid = true
		}

		// 2. 检查配置文件secret
		if !valid && secretKey != "" {
			// 如果是bcrypt哈希,使用bcrypt验证
			if strings.HasPrefix(secretKey, "$2") {
				if err := bcrypt.CompareHashAndPassword([]byte(secretKey), []byte(token)); err == nil {
					valid = true
				}
			} else {
				// 明文比较(不推荐,仅用于开发)
				if subtle.ConstantTimeCompare([]byte(token), []byte(secretKey)) == 1 {
					valid = true
				}
			}
		}

		// 3. 检查本地密码(仅本地客户端)
		if !valid && localClient && a.config.LocalPassword != "" {
			if subtle.ConstantTimeCompare([]byte(token), []byte(a.config.LocalPassword)) == 1 {
				valid = true
			}
		}

		if !valid {
			if !localClient {
				a.recordFailure(clientIP)
			}
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的管理员API密钥",
				"error":   "invalid_token",
			})
			c.Abort()
			return
		}

		// 认证成功,清除失败记录
		if !localClient {
			a.clearFailure(clientIP)
		}

		c.Next()
	}
}

// extractToken 从请求中提取认证token
// 支持以下方式:
// 1. Authorization: Bearer <token>
// 2. X-Admin-API-Key: <token>
func (a *AdminAuth) extractToken(c *gin.Context) string {
	// 尝试Authorization header
	if auth := c.GetHeader("Authorization"); auth != "" {
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return strings.TrimSpace(parts[1])
		}
		// 也支持直接传递token(不带Bearer前缀)
		return strings.TrimSpace(auth)
	}

	// 尝试X-Admin-API-Key header
	if key := c.GetHeader("X-Admin-API-Key"); key != "" {
		return strings.TrimSpace(key)
	}

	return ""
}

// recordFailure 记录失败尝试
func (a *AdminAuth) recordFailure(clientIP string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	const maxFailures = 5
	const banDuration = 30 * time.Minute

	ai := a.failedAttempts[clientIP]
	if ai == nil {
		ai = &attemptInfo{}
		a.failedAttempts[clientIP] = ai
	}

	ai.count++
	ai.lastActivity = time.Now()

	if ai.count >= maxFailures {
		ai.blockedUntil = time.Now().Add(banDuration)
		ai.count = 0
	}
}

// clearFailure 清除失败记录
func (a *AdminAuth) clearFailure(clientIP string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if ai := a.failedAttempts[clientIP]; ai != nil {
		ai.count = 0
		ai.blockedUntil = time.Time{}
	}
}

// AdminAuthMiddleware 简化的管理员认证中间件(全局单例)
var adminAuth *AdminAuth

// InitAdminAuth 初始化管理员认证
func InitAdminAuth(secretKey string, allowRemote bool) {
	adminAuth = NewAdminAuth(secretKey, allowRemote)
}

// AdminAuthMiddleware 返回管理员认证中间件(便捷函数)
func AdminAuthMiddleware() gin.HandlerFunc {
	if adminAuth == nil {
		// 未初始化,返回允许所有请求的中间件
		return func(c *gin.Context) {
			c.Next()
		}
	}
	return adminAuth.Middleware()
}

// OptionalAdminAuth 可选的管理员认证
// 如果提供了token则验证,未提供则跳过
func OptionalAdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果未配置密钥,跳过认证
		if adminAuth == nil {
			c.Next()
			return
		}

		secretKey := adminAuth.config.SecretKey
		envSecret := adminAuth.envSecret

		if secretKey == "" && envSecret == "" {
			c.Next()
			return
		}

		// 检查是否提供了认证信息
		token := adminAuth.extractToken(c)
		if token == "" {
			// 未提供认证信息,标记为未认证但继续执行
			c.Set("authenticated", false)
			c.Next()
			return
		}

		// 验证token
		valid := false

		if envSecret != "" && subtle.ConstantTimeCompare([]byte(token), []byte(envSecret)) == 1 {
			valid = true
		} else if secretKey != "" {
			if strings.HasPrefix(secretKey, "$2") {
				if err := bcrypt.CompareHashAndPassword([]byte(secretKey), []byte(token)); err == nil {
					valid = true
				}
			} else if subtle.ConstantTimeCompare([]byte(token), []byte(secretKey)) == 1 {
				valid = true
			}
		}

		c.Set("authenticated", valid)
		c.Next()
	}
}
