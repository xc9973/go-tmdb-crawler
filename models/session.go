package models

import (
	"fmt"
	"time"
)

// Session represents an auth session stored in the database
// Token is stored as raw JWT string for simple lookup and invalidation.
type Session struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	SessionID  string    `gorm:"size:64;uniqueIndex:idx_session_id;not null" json:"session_id"`
	Token      string    `gorm:"type:text;not null" json:"token"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	ExpiresAt  time.Time `gorm:"index:idx_expires_at;not null" json:"expires_at"`
	LastActive time.Time `gorm:"autoUpdateTime" json:"last_active"`
	UserAgent  string    `gorm:"type:text" json:"user_agent"`
	IP         string    `gorm:"size:64;index:idx_ip" json:"ip"`
}

// TableName specifies the table name for Session model
func (Session) TableName() string {
	return "sessions"
}

// Validate validates the session data
func (s *Session) Validate() error {
	if s.SessionID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}
	if s.Token == "" {
		return fmt.Errorf("token cannot be empty")
	}
	if s.ExpiresAt.IsZero() {
		return fmt.Errorf("expiration time cannot be empty")
	}
	return nil
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsValid checks if the session is still valid
func (s *Session) IsValid() bool {
	return !s.IsExpired()
}

// Refresh extends the session expiration
func (s *Session) Refresh(duration time.Duration) {
	s.ExpiresAt = time.Now().Add(duration)
	s.LastActive = time.Now()
}
