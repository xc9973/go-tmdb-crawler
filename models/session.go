package models

import "time"

// Session represents an auth session stored in the database
// Token is stored as raw JWT string for simple lookup and invalidation.
type Session struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	SessionID  string    `gorm:"size:64;uniqueIndex" json:"session_id"`
	Token      string    `gorm:"type:text" json:"token"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `gorm:"index" json:"expires_at"`
	LastActive time.Time `json:"last_active"`
	UserAgent  string    `gorm:"type:text" json:"user_agent"`
	IP         string    `gorm:"size:64" json:"ip"`
}

// TableName specifies the table name for Session model
func (Session) TableName() string {
	return "sessions"
}
