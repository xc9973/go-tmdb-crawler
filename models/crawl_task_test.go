package models

import (
	"testing"
	"time"
)

func TestCrawlTask_Validate(t *testing.T) {
	tests := []struct {
		name    string
		task    *CrawlTask
		wantErr bool
	}{
		{
			name: "Valid refresh_all task",
			task: &CrawlTask{
				Type:   "refresh_all",
				Status: "queued",
			},
			wantErr: false,
		},
		{
			name: "Valid crawl_by_id task",
			task: &CrawlTask{
				Type:   "crawl_by_id",
				Status: "running",
			},
			wantErr: false,
		},
		{
			name: "Valid crawl_by_status task",
			task: &CrawlTask{
				Type:   "crawl_by_status",
				Status: "success",
			},
			wantErr: false,
		},
		{
			name: "Valid daily_job task",
			task: &CrawlTask{
				Type:   "daily_job",
				Status: "failed",
			},
			wantErr: false,
		},
		{
			name: "Invalid task type",
			task: &CrawlTask{
				Type:   "invalid_type",
				Status: "queued",
			},
			wantErr: true,
		},
		{
			name: "Invalid task status",
			task: &CrawlTask{
				Type:   "refresh_all",
				Status: "invalid_status",
			},
			wantErr: true,
		},
		{
			name: "Empty task type",
			task: &CrawlTask{
				Type:   "",
				Status: "queued",
			},
			wantErr: true,
		},
		{
			name: "Empty task status",
			task: &CrawlTask{
				Type:   "refresh_all",
				Status: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.task.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("CrawlTask.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCrawlTask_IsRunning(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{
			name:     "Running status",
			status:   "running",
			expected: true,
		},
		{
			name:     "Queued status",
			status:   "queued",
			expected: false,
		},
		{
			name:     "Success status",
			status:   "success",
			expected: false,
		},
		{
			name:     "Failed status",
			status:   "failed",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &CrawlTask{Status: tt.status}
			if got := task.IsRunning(); got != tt.expected {
				t.Errorf("CrawlTask.IsRunning() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCrawlTask_IsCompleted(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{
			name:     "Success status",
			status:   "success",
			expected: true,
		},
		{
			name:     "Failed status",
			status:   "failed",
			expected: true,
		},
		{
			name:     "Running status",
			status:   "running",
			expected: false,
		},
		{
			name:     "Queued status",
			status:   "queued",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &CrawlTask{Status: tt.status}
			if got := task.IsCompleted(); got != tt.expected {
				t.Errorf("CrawlTask.IsCompleted() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCrawlTask_GetDuration(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		startedAt   *time.Time
		finishedAt  *time.Time
		wantNil     bool
		description string
	}{
		{
			name:        "Both timestamps present",
			startedAt:   timePtr(now.Add(-1 * time.Hour)),
			finishedAt:  timePtr(now),
			wantNil:     false,
			description: "Should return duration",
		},
		{
			name:        "Only started timestamp",
			startedAt:   timePtr(now.Add(-1 * time.Hour)),
			finishedAt:  nil,
			wantNil:     true,
			description: "Should return nil when not finished",
		},
		{
			name:        "Only finished timestamp",
			startedAt:   nil,
			finishedAt:  timePtr(now),
			wantNil:     true,
			description: "Should return nil when not started",
		},
		{
			name:        "No timestamps",
			startedAt:   nil,
			finishedAt:  nil,
			wantNil:     true,
			description: "Should return nil when no timestamps",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &CrawlTask{
				StartedAt:  tt.startedAt,
				FinishedAt: tt.finishedAt,
			}
			got := task.GetDuration()
			if (got == nil) != tt.wantNil {
				t.Errorf("CrawlTask.GetDuration() = %v, wantNil %v (%s)", got, tt.wantNil, tt.description)
			}
			if !tt.wantNil && got != nil {
				// Verify duration is positive
				if *got <= 0 {
					t.Errorf("CrawlTask.GetDuration() returned non-positive duration: %v", got)
				}
			}
		})
	}
}

func TestCrawlTask_TableName(t *testing.T) {
	task := CrawlTask{}
	if got := task.TableName(); got != "crawl_tasks" {
		t.Errorf("CrawlTask.TableName() = %v, want %v", got, "crawl_tasks")
	}
}

func TestCrawlTask_GetDuration_Accuracy(t *testing.T) {
	now := time.Now()
	started := now.Add(-10 * time.Minute)
	finished := now.Add(-5 * time.Minute)

	task := &CrawlTask{
		StartedAt:  timePtr(started),
		FinishedAt: timePtr(finished),
	}

	duration := task.GetDuration()
	if duration == nil {
		t.Fatal("Expected non-nil duration")
	}

	expected := 5 * time.Minute
	if *duration != expected {
		t.Errorf("CrawlTask.GetDuration() = %v, want %v", *duration, expected)
	}
}
