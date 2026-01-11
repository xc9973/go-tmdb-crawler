package models

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"
)

// TelegraphPost represents a published Telegraph article
type TelegraphPost struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Title         string    `gorm:"size:255;not null" json:"title"`
	TelegraphURL  string    `gorm:"size:512" json:"telegraph_url"`
	TelegraphPath string    `gorm:"size:255" json:"telegraph_path"`
	ContentHash   string    `gorm:"size:64;index" json:"content_hash"` // MD5 hash
	ShowsCount    int       `json:"shows_count"`
	EpisodesCount int       `json:"episodes_count"`
	DateRange     string    `gorm:"size:50" json:"date_range"` // '2026-01-11 to 2026-02-10'
	CreatedAt     time.Time `gorm:"index" json:"created_at"`
}

// TableName specifies the table name for TelegraphPost model
func (TelegraphPost) TableName() string {
	return "telegraph_posts"
}

// GetFullURL returns the full Telegraph URL
func (t *TelegraphPost) GetFullURL() string {
	if t.TelegraphURL != "" {
		return t.TelegraphURL
	}
	return fmt.Sprintf("https://telegra.ph/%s", t.TelegraphPath)
}

// GenerateContentHash generates MD5 hash of content
func GenerateContentHash(content string) string {
	hash := md5.Sum([]byte(content))
	return hex.EncodeToString(hash[:])
}

// IsRecent checks if the post was created within the last 24 hours
func (t *TelegraphPost) IsRecent() bool {
	return time.Since(t.CreatedAt) < 24*time.Hour
}
