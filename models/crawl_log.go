package models

import (
	"fmt"
	"time"
)

// CrawlLog represents a crawling operation log
type CrawlLog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ShowID        *uint     `gorm:"index:idx_log_show_id" json:"show_id,omitempty"`
	TmdbID        int       `gorm:"not null;index:idx_log_tmdb_id" json:"tmdb_id"`
	Action        string    `gorm:"size:50;not null;index:idx_log_action" json:"action"`                 // 'fetch'/'refresh'/'batch'
	Status        string    `gorm:"size:20;not null;index:idx_log_status;default:success" json:"status"` // 'success'/'failed'/'partial'
	EpisodesCount int       `gorm:"default:0" json:"episodes_count"`
	ErrorMessage  string    `gorm:"type:text" json:"error_message,omitempty"`
	DurationMs    int       `gorm:"default:0" json:"duration_ms"` // Duration in milliseconds
	CreatedAt     time.Time `gorm:"index:idx_log_created_at;autoCreateTime" json:"created_at"`

	// Relationships
	Show *Show `gorm:"foreignKey:ShowID;constraint:OnDelete:SET NULL" json:"show,omitempty"`
}

// TableName specifies the table name for CrawlLog model
func (CrawlLog) TableName() string {
	return "crawl_logs"
}

// IsSuccess checks if the crawl was successful
func (c *CrawlLog) IsSuccess() bool {
	return c.Status == "success"
}

// IsFailed checks if the crawl failed
func (c *CrawlLog) IsFailed() bool {
	return c.Status == "failed"
}

// IsPartial checks if the crawl was partially successful
func (c *CrawlLog) IsPartial() bool {
	return c.Status == "partial"
}

// GetDuration returns the duration as a formatted string
func (c *CrawlLog) GetDuration() string {
	if c.DurationMs < 1000 {
		return fmt.Sprintf("%dms", c.DurationMs)
	}
	seconds := c.DurationMs / 1000
	return fmt.Sprintf("%ds", seconds)
}

// Validate validates the crawl log data
func (c *CrawlLog) Validate() error {
	if c.TmdbID <= 0 {
		return fmt.Errorf("invalid TMDB ID")
	}

	validActions := map[string]bool{
		"fetch":   true,
		"refresh": true,
		"batch":   true,
	}
	if !validActions[c.Action] {
		return fmt.Errorf("invalid action: %s", c.Action)
	}

	validStatuses := map[string]bool{
		"success": true,
		"failed":  true,
		"partial": true,
	}
	if !validStatuses[c.Status] {
		return fmt.Errorf("invalid status: %s", c.Status)
	}

	return nil
}

// GetStatusIcon returns an emoji icon based on status
func (c *CrawlLog) GetStatusIcon() string {
	switch c.Status {
	case "success":
		return "✅"
	case "failed":
		return "❌"
	case "partial":
		return "⚠️"
	default:
		return "❓"
	}
}
