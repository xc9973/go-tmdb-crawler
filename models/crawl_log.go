package models

import (
	"fmt"
	"time"
)

// CrawlLog represents a crawling operation log
type CrawlLog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ShowID        *uint     `gorm:"index" json:"show_id,omitempty"`
	TmdbID        int       `json:"tmdb_id"`
	Action        string    `gorm:"size:50" json:"action"`       // 'fetch'/'refresh'/'batch'
	Status        string    `gorm:"size:20;index" json:"status"` // 'success'/'failed'/'partial'
	EpisodesCount int       `json:"episodes_count"`
	ErrorMessage  string    `gorm:"type:text" json:"error_message,omitempty"`
	DurationMs    int       `json:"duration_ms"` // Duration in milliseconds
	CreatedAt     time.Time `gorm:"index" json:"created_at"`

	// Relationships
	Show *Show `gorm:"foreignKey:ShowID" json:"show,omitempty"`
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
