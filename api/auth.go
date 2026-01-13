package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xc9973/go-tmdb-crawler/services"
)

// AdminAuthValidator 管理员API Key验证器接口
type AdminAuthValidator interface {
	ValidateAPIKey(apiKey string) bool
}

// AuthHandler 认证处理器
type AuthHandler struct {
	authService *services.AuthService
	adminAuth   AdminAuthValidator
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService *services.AuthService, adminAuth AdminAuthValidator) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		adminAuth:   adminAuth,
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	APIKey     string `json:"api_key" binding:"required"`
	RememberMe bool   `json:"remember_me"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token        string    `json:"token"`
	ExpiresAt    time.Time `json:"expires_at"`
	SessionID    string    `json:"session_id"`
	IsFirstLogin bool      `json:"is_first_login"`
	Message      string    `json:"message,omitempty"`
}

// RefreshTokenResponse 刷新token响应
type RefreshTokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	SessionID string    `json:"session_id"`
}

// Login 登录接口
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 验证API Key
	if !h.adminAuth.ValidateAPIKey(req.APIKey) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "API密钥无效",
			"error":   "invalid_api_key",
		})
		return
	}

	// 生成token
	userAgent := c.GetHeader("User-Agent")
	ip := c.ClientIP()

	token, session, err := h.authService.Login(req.APIKey, userAgent, ip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "登录失败",
			"error":   err.Error(),
		})
		return
	}

	// 计算cookie过期时间
	// 如果选择"记住我",则使用session的过期时间(30天)
	// 否则使用会话cookie(浏览器关闭后失效)
	maxAge := int(session.ExpiresAt.Sub(time.Now()).Seconds())
	if !req.RememberMe {
		maxAge = 0 // 会话cookie
	}

	// 设置httpOnly cookie
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		"session_token",
		token,
		maxAge,
		"/",
		"",
		false, // secure (生产环境应为true,需要HTTPS)
		true,  // httpOnly - 防止XSS攻击
	)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登录成功",
		"success": true,
		"data": LoginResponse{
			Token:        token,
			ExpiresAt:    session.ExpiresAt,
			SessionID:    extractSessionID(token),
			IsFirstLogin: true, // 首次登录标识
			Message:      "登录成功",
		},
	})
}

// Logout 登出接口
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// 从cookie中获取token
	token, err := c.Cookie("session_token")
	if err == nil && token != "" {
		// 删除session
		h.authService.Logout(token)
	}

	// 清除cookie
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		"session_token",
		"",
		-1,
		"/",
		"",
		false,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登出成功",
	})
}

// RefreshToken 刷新token接口
// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// 从cookie中获取token
	token, err := c.Cookie("session_token")
	if err != nil || token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未登录",
			"error":   "not_authenticated",
		})
		return
	}

	// 刷新token
	newToken, session, err := h.authService.RefreshToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "Token无效或已过期",
			"error":   err.Error(),
		})
		return
	}

	// 更新cookie
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		"session_token",
		newToken,
		int(session.ExpiresAt.Sub(time.Now()).Seconds()),
		"/",
		"",
		false,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Token刷新成功",
		"data": RefreshTokenResponse{
			Token:     newToken,
			ExpiresAt: session.ExpiresAt,
			SessionID: extractSessionID(newToken),
		},
	})
}

// GetSessionInfo 获取当前session信息
// GET /api/v1/auth/session
func (h *AuthHandler) GetSessionInfo(c *gin.Context) {
	// 从cookie中获取token
	token, err := c.Cookie("session_token")
	if err != nil || token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未登录",
			"error":   "not_authenticated",
			"data": gin.H{
				"authenticated": false,
			},
		})
		return
	}

	// 验证token
	sessionInterface, err := h.authService.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "Token无效或已过期",
			"error":   err.Error(),
			"data": gin.H{
				"authenticated": false,
			},
		})
		return
	}

	session := sessionInterface.(*services.SessionInfo)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"authenticated": true,
			"session_id":    extractSessionID(token),
			"created_at":    session.CreatedAt,
			"expires_at":    session.ExpiresAt,
			"last_active":   session.LastActive,
			"user_agent":    session.UserAgent,
			"ip":            session.IP,
		},
	})
}

// extractSessionID 从token中提取session ID (简化版)
func extractSessionID(token string) string {
	// JWT token格式: header.payload.signature
	// 这里简化处理，实际应该解析JWT
	parts := strings.Split(token, ".")
	if len(parts) >= 2 {
		return parts[1][:16] // 返回payload的前16个字符作为简化的session ID
	}
	return ""
}
