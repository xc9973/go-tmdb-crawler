package models

import (
	"fmt"
	"time"
)

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
	Type         string     `gorm:"size:50;not null;index:idx_task_type" json:"type"`
	Status       string     `gorm:"size:20;not null;index:idx_task_status;default:queued" json:"status"`
	Params       string     `gorm:"type:text" json:"params,omitempty"`
	ErrorMessage string     `gorm:"type:text" json:"error_message,omitempty"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
	CreatedAt    time.Time  `gorm:"index:idx_created_at;autoCreateTime" json:"created_at"`
}

// TableName specifies the table name for CrawlTask model
func (CrawlTask) TableName() string {
	return "crawl_tasks"
}

// Validate validates the crawl task data
func (c *CrawlTask) Validate() error {
	validTypes := map[string]bool{
		"refresh_all":     true,
		"crawl_by_id":     true,
		"crawl_by_status": true,
		"daily_job":       true,
	}
	if !validTypes[c.Type] {
		return fmt.Errorf("invalid task type: %s", c.Type)
	}

	validStatuses := map[string]bool{
		"queued":  true,
		"running": true,
		"success": true,
		"failed":  true,
	}
	if !validStatuses[c.Status] {
		return fmt.Errorf("invalid task status: %s", c.Status)
	}

	return nil
}

// IsRunning checks if the task is currently running
func (c *CrawlTask) IsRunning() bool {
	return c.Status == "running"
}

// IsCompleted checks if the task has completed (success or failed)
func (c *CrawlTask) IsCompleted() bool {
	return c.Status == "success" || c.Status == "failed"
}

// GetDuration returns the task execution duration
func (c *CrawlTask) GetDuration() *time.Duration {
	if c.StartedAt == nil || c.FinishedAt == nil {
		return nil
	}
	duration := c.FinishedAt.Sub(*c.StartedAt)
	return &duration
}
