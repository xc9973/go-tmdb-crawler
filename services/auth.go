package services

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/xc9973/go-tmdb-crawler/models"
	"gorm.io/gorm"
)

// AuthService 认证服务
type AuthService struct {
	secretKey       string
	sessionDuration time.Duration
	db              *gorm.DB
	sessions        map[string]*SessionInfo
	mu              sync.RWMutex
}

// SessionInfo 会话信息
type SessionInfo struct {
	Token      string
	CreatedAt  time.Time
	ExpiresAt  time.Time
	LastActive time.Time
	UserAgent  string
	IP         string
}

// SessionInfoAlias SessionInfo的别名，用于导出
type SessionInfoAlias = SessionInfo

// JWTClaims JWT声明
type JWTClaims struct {
	SessionID string `json:"session_id"`
	jwt.RegisteredClaims
}

// NewAuthService 创建认证服务
func NewAuthService(secretKey string, db *gorm.DB) *AuthService {
	return &AuthService{
		secretKey:       secretKey,
		sessionDuration: 2 * time.Hour, // 默认2小时
		db:              db,
		sessions:        make(map[string]*SessionInfo),
	}
}

// Login 验证API Key并生成token
func (s *AuthService) Login(apiKey string, userAgent, ip string) (string, *SessionInfo, error) {
	if apiKey == "" {
		return "", nil, errors.New("API key is required")
	}

	// 生成session ID
	sessionID, err := s.generateSessionID()
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	// 生成JWT token
	token, err := s.generateToken(sessionID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// 创建session
	now := time.Now()
	session := &SessionInfo{
		Token:      token,
		CreatedAt:  now,
		ExpiresAt:  now.Add(s.sessionDuration),
		LastActive: now,
		UserAgent:  userAgent,
		IP:         ip,
	}

	// 保存session
	s.mu.Lock()
	s.sessions[sessionID] = session
	s.mu.Unlock()

	if s.db != nil {
		record := &models.Session{
			SessionID:  sessionID,
			Token:      token,
			CreatedAt:  session.CreatedAt,
			ExpiresAt:  session.ExpiresAt,
			LastActive: session.LastActive,
			UserAgent:  session.UserAgent,
			IP:         session.IP,
		}
		if err := s.db.Create(record).Error; err != nil {
			return "", nil, fmt.Errorf("failed to persist session: %w", err)
		}
	}

	return token, session, nil
}

// ValidateToken 验证token并返回session信息
// 实现middleware.AuthService接口
func (s *AuthService) ValidateToken(tokenString string) (interface{}, error) {
	// 解析JWT
	claims := &JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("token is invalid")
	}

	// 检查session是否存在（先内存，再DB）
	s.mu.RLock()
	session, exists := s.sessions[claims.SessionID]
	s.mu.RUnlock()

	if !exists && s.db != nil {
		var record models.Session
		if err := s.db.Where("session_id = ? AND token = ?", claims.SessionID, tokenString).First(&record).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("session not found")
			}
			return nil, fmt.Errorf("failed to load session: %w", err)
		}
		session = &SessionInfo{
			Token:      record.Token,
			CreatedAt:  record.CreatedAt,
			ExpiresAt:  record.ExpiresAt,
			LastActive: record.LastActive,
			UserAgent:  record.UserAgent,
			IP:         record.IP,
		}
		exists = true
		s.mu.Lock()
		s.sessions[claims.SessionID] = session
		s.mu.Unlock()
	}

	if !exists {
		return nil, errors.New("session not found")
	}

	// 检查是否过期
	if time.Now().After(session.ExpiresAt) {
		s.mu.Lock()
		delete(s.sessions, claims.SessionID)
		s.mu.Unlock()
		if s.db != nil {
			_ = s.db.Where("session_id = ?", claims.SessionID).Delete(&models.Session{}).Error
		}
		return nil, errors.New("session expired")
	}

	// 更新最后活跃时间
	s.mu.Lock()
	session.LastActive = time.Now()
	s.mu.Unlock()

	if s.db != nil {
		_ = s.db.Model(&models.Session{}).
			Where("session_id = ?", claims.SessionID).
			Update("last_active", session.LastActive).Error
	}

	return session, nil
}

// Logout 登出并删除session
func (s *AuthService) Logout(tokenString string) error {
	// 解析JWT获取session ID
	claims := &JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.secretKey), nil
	})

	if err != nil || !token.Valid {
		return errors.New("invalid token")
	}

	// 删除session
	s.mu.Lock()
	delete(s.sessions, claims.SessionID)
	s.mu.Unlock()

	if s.db != nil {
		if err := s.db.Where("session_id = ?", claims.SessionID).Delete(&models.Session{}).Error; err != nil {
			return fmt.Errorf("failed to delete session: %w", err)
		}
	}

	return nil
}

// RefreshToken 刷新token
func (s *AuthService) RefreshToken(tokenString string) (string, *SessionInfo, error) {
	// 验证旧token
	sessionInterface, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", nil, err
	}

	session := sessionInterface.(*SessionInfo)

	// 生成新的session ID和token
	sessionID, err := s.generateSessionID()
	if err != nil {
		return "", nil, err
	}

	newToken, err := s.generateToken(sessionID)
	if err != nil {
		return "", nil, err
	}

	// 更新session
	now := time.Now()
	s.mu.Lock()
	// 删除旧session
	oldClaims := &JWTClaims{}
	if oldToken, _ := jwt.ParseWithClaims(tokenString, oldClaims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.secretKey), nil
	}); oldToken != nil {
		delete(s.sessions, oldClaims.SessionID)
	}

	// 创建新session
	newSession := &SessionInfo{
		Token:      newToken,
		CreatedAt:  session.CreatedAt,
		ExpiresAt:  now.Add(s.sessionDuration),
		LastActive: now,
		UserAgent:  session.UserAgent,
		IP:         session.IP,
	}
	s.sessions[sessionID] = newSession
	s.mu.Unlock()

	if s.db != nil {
		if oldClaims.SessionID != "" {
			_ = s.db.Where("session_id = ?", oldClaims.SessionID).Delete(&models.Session{}).Error
		}
		record := &models.Session{
			SessionID:  sessionID,
			Token:      newToken,
			CreatedAt:  newSession.CreatedAt,
			ExpiresAt:  newSession.ExpiresAt,
			LastActive: newSession.LastActive,
			UserAgent:  newSession.UserAgent,
			IP:         newSession.IP,
		}
		if err := s.db.Create(record).Error; err != nil {
			return "", nil, fmt.Errorf("failed to persist session: %w", err)
		}
	}

	return newToken, newSession, nil
}

// CleanupExpiredSessions 清理过期session
func (s *AuthService) CleanupExpiredSessions() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for sessionID, session := range s.sessions {
		if now.After(session.ExpiresAt) {
			delete(s.sessions, sessionID)
		}
	}

	if s.db != nil {
		_ = s.db.Where("expires_at < ?", now).Delete(&models.Session{}).Error
	}
}

// GetActiveSessionCount 获取活跃session数量
func (s *AuthService) GetActiveSessionCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	now := time.Now()
	for _, session := range s.sessions {
		if now.Before(session.ExpiresAt) {
			count++
		}
	}

	if s.db == nil {
		return count
	}

	var dbCount int64
	if err := s.db.Model(&models.Session{}).Where("expires_at > ?", now).Count(&dbCount).Error; err == nil {
		return int(dbCount)
	}

	return count
}

// generateToken 生成JWT token
func (s *AuthService) generateToken(sessionID string) (string, error) {
	claims := JWTClaims{
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.sessionDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secretKey))
}

// generateSessionID 生成随机的session ID
func (s *AuthService) generateSessionID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
