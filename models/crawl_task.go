package models

import "time"

// CrawlTask represents an async crawl task
// Status: queued/running/success/failed
// Type: refresh_all/crawl_by_status
// Params: JSON string for task inputs
// ErrorMessage: failure reason, if any
// StartedAt/FinishedAt: timestamps for execution window
//
// Note: keep fields minimal to avoid schema churn.
type CrawlTask struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Type         string     `gorm:"size:50" json:"type"`
	Status       string     `gorm:"size:20;index" json:"status"`
	Params       string     `gorm:"type:text" json:"params,omitempty"`
	ErrorMessage string     `gorm:"type:text" json:"error_message,omitempty"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
	CreatedAt    time.Time  `gorm:"index" json:"created_at"`
}

// TableName specifies the table name for CrawlTask model
func (CrawlTask) TableName() string {
	return "crawl_tasks"
}
