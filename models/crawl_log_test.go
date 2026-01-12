package models

import (
	"testing"
)

func TestCrawlLog_IsSuccess(t *testing.T) {
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
			expected: false,
		},
		{
			name:     "Partial status",
			status:   "partial",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := &CrawlLog{Status: tt.status}
			if got := log.IsSuccess(); got != tt.expected {
				t.Errorf("CrawlLog.IsSuccess() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCrawlLog_IsFailed(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{
			name:     "Failed status",
			status:   "failed",
			expected: true,
		},
		{
			name:     "Success status",
			status:   "success",
			expected: false,
		},
		{
			name:     "Partial status",
			status:   "partial",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := &CrawlLog{Status: tt.status}
			if got := log.IsFailed(); got != tt.expected {
				t.Errorf("CrawlLog.IsFailed() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCrawlLog_IsPartial(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{
			name:     "Partial status",
			status:   "partial",
			expected: true,
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
			log := &CrawlLog{Status: tt.status}
			if got := log.IsPartial(); got != tt.expected {
				t.Errorf("CrawlLog.IsPartial() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCrawlLog_GetDuration(t *testing.T) {
	tests := []struct {
		name       string
		durationMs int
		expected   string
	}{
		{
			name:       "Less than 1 second",
			durationMs: 500,
			expected:   "500ms",
		},
		{
			name:       "Exactly 1 second",
			durationMs: 1000,
			expected:   "1s",
		},
		{
			name:       "More than 1 second",
			durationMs: 2500,
			expected:   "2s",
		},
		{
			name:       "Zero duration",
			durationMs: 0,
			expected:   "0ms",
		},
		{
			name:       "Large duration",
			durationMs: 65000,
			expected:   "65s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := &CrawlLog{DurationMs: tt.durationMs}
			if got := log.GetDuration(); got != tt.expected {
				t.Errorf("CrawlLog.GetDuration() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCrawlLog_Validate(t *testing.T) {
	tests := []struct {
		name    string
		log     *CrawlLog
		wantErr bool
	}{
		{
			name: "Valid fetch log",
			log: &CrawlLog{
				TmdbID: 123,
				Action: "fetch",
				Status: "success",
			},
			wantErr: false,
		},
		{
			name: "Valid refresh log",
			log: &CrawlLog{
				TmdbID: 456,
				Action: "refresh",
				Status: "failed",
			},
			wantErr: false,
		},
		{
			name: "Valid batch log",
			log: &CrawlLog{
				TmdbID: 789,
				Action: "batch",
				Status: "partial",
			},
			wantErr: false,
		},
		{
			name: "Invalid TMDB ID",
			log: &CrawlLog{
				TmdbID: 0,
				Action: "fetch",
				Status: "success",
			},
			wantErr: true,
		},
		{
			name: "Negative TMDB ID",
			log: &CrawlLog{
				TmdbID: -1,
				Action: "fetch",
				Status: "success",
			},
			wantErr: true,
		},
		{
			name: "Invalid action",
			log: &CrawlLog{
				TmdbID: 123,
				Action: "invalid_action",
				Status: "success",
			},
			wantErr: true,
		},
		{
			name: "Invalid status",
			log: &CrawlLog{
				TmdbID: 123,
				Action: "fetch",
				Status: "invalid_status",
			},
			wantErr: true,
		},
		{
			name: "Empty action",
			log: &CrawlLog{
				TmdbID: 123,
				Action: "",
				Status: "success",
			},
			wantErr: true,
		},
		{
			name: "Empty status",
			log: &CrawlLog{
				TmdbID: 123,
				Action: "fetch",
				Status: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.log.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("CrawlLog.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCrawlLog_GetStatusIcon(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected string
	}{
		{
			name:     "Success status",
			status:   "success",
			expected: "✅",
		},
		{
			name:     "Failed status",
			status:   "failed",
			expected: "❌",
		},
		{
			name:     "Partial status",
			status:   "partial",
			expected: "⚠️",
		},
		{
			name:     "Unknown status",
			status:   "unknown",
			expected: "❓",
		},
		{
			name:     "Empty status",
			status:   "",
			expected: "❓",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := &CrawlLog{Status: tt.status}
			if got := log.GetStatusIcon(); got != tt.expected {
				t.Errorf("CrawlLog.GetStatusIcon() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCrawlLog_TableName(t *testing.T) {
	log := CrawlLog{}
	if got := log.TableName(); got != "crawl_logs" {
		t.Errorf("CrawlLog.TableName() = %v, want %v", got, "crawl_logs")
	}
}
