package models

import (
	"testing"
	"time"
)

func TestShow_IsReturning(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{
			name:     "Returning Series",
			status:   "Returning Series",
			expected: true,
		},
		{
			name:     "Ended",
			status:   "Ended",
			expected: false,
		},
		{
			name:     "Canceled",
			status:   "Canceled",
			expected: false,
		},
		{
			name:     "Empty status",
			status:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			show := &Show{Status: tt.status}
			if got := show.IsReturning(); got != tt.expected {
				t.Errorf("Show.IsReturning() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestShow_IsEnded(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{
			name:     "Ended",
			status:   "Ended",
			expected: true,
		},
		{
			name:     "Returning Series",
			status:   "Returning Series",
			expected: false,
		},
		{
			name:     "Canceled",
			status:   "Canceled",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			show := &Show{Status: tt.status}
			if got := show.IsEnded(); got != tt.expected {
				t.Errorf("Show.IsEnded() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestShow_GetDisplayStatus(t *testing.T) {
	tests := []struct {
		name         string
		status       string
		customStatus string
		expected     string
	}{
		{
			name:         "Custom status takes precedence",
			status:       "Returning Series",
			customStatus: "Watching",
			expected:     "Watching",
		},
		{
			name:         "No custom status",
			status:       "Returning Series",
			customStatus: "",
			expected:     "Returning Series",
		},
		{
			name:         "Both empty",
			status:       "",
			customStatus: "",
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			show := &Show{
				Status:       tt.status,
				CustomStatus: tt.customStatus,
			}
			if got := show.GetDisplayStatus(); got != tt.expected {
				t.Errorf("Show.GetDisplayStatus() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestShow_GetDisplayType(t *testing.T) {
	tests := []struct {
		name     string
		showType string
		expected string
	}{
		{
			name:     "Scripted",
			showType: "Scripted",
			expected: "Scripted",
		},
		{
			name:     "Reality",
			showType: "Reality",
			expected: "Reality",
		},
		{
			name:     "Empty type",
			showType: "",
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			show := &Show{Type: tt.showType}
			if got := show.GetDisplayType(); got != tt.expected {
				t.Errorf("Show.GetDisplayType() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestShow_Validate(t *testing.T) {
	tests := []struct {
		name    string
		show    *Show
		wantErr bool
	}{
		{
			name: "Valid show",
			show: &Show{
				Name:   "Test Show",
				TmdbID: 123,
			},
			wantErr: false,
		},
		{
			name: "Empty name",
			show: &Show{
				Name:   "",
				TmdbID: 123,
			},
			wantErr: true,
		},
		{
			name: "Invalid TMDB ID",
			show: &Show{
				Name:   "Test Show",
				TmdbID: 0,
			},
			wantErr: true,
		},
		{
			name: "Negative TMDB ID",
			show: &Show{
				Name:   "Test Show",
				TmdbID: -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.show.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Show.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestShow_IsExpired(t *testing.T) {
	tests := []struct {
		name        string
		lastCrawled *time.Time
		expected    bool
		description string
	}{
		{
			name:        "Never crawled",
			lastCrawled: nil,
			expected:    true,
			description: "Show with no crawl time should be expired",
		},
		{
			name:        "Crawled recently",
			lastCrawled: timePtr(time.Now().Add(-1 * time.Hour)),
			expected:    false,
			description: "Show crawled 1 hour ago should not be expired",
		},
		{
			name:        "Crawled 25 hours ago",
			lastCrawled: timePtr(time.Now().Add(-25 * time.Hour)),
			expected:    true,
			description: "Show crawled 25 hours ago should be expired",
		},
		{
			name:        "Crawled exactly 24 hours ago",
			lastCrawled: timePtr(time.Now().Add(-24*time.Hour + 1*time.Second)),
			expected:    false,
			description: "Show crawled just under 24 hours ago should not be expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			show := &Show{LastCrawledAt: tt.lastCrawled}
			if got := show.IsExpired(); got != tt.expected {
				t.Errorf("Show.IsExpired() = %v, want %v (%s)", got, tt.expected, tt.description)
			}
		})
	}
}

func TestShow_ShouldRefresh(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		status      string
		lastCrawled *time.Time
		expected    bool
		description string
	}{
		{
			name:        "Returning series - never crawled",
			status:      "Returning Series",
			lastCrawled: nil,
			expected:    true,
			description: "Returning series with no crawl time should refresh",
		},
		{
			name:        "Returning series - crawled recently",
			status:      "Returning Series",
			lastCrawled: timePtr(now.Add(-1 * time.Hour)),
			expected:    false,
			description: "Returning series crawled recently should not refresh",
		},
		{
			name:        "Returning series - crawled 25 hours ago",
			status:      "Returning Series",
			lastCrawled: timePtr(now.Add(-25 * time.Hour)),
			expected:    true,
			description: "Returning series crawled 25 hours ago should refresh",
		},
		{
			name:        "Ended series - never crawled",
			status:      "Ended",
			lastCrawled: nil,
			expected:    true,
			description: "Ended series with no crawl time should refresh",
		},
		{
			name:        "Ended series - crawled 1 day ago",
			status:      "Ended",
			lastCrawled: timePtr(now.Add(-24 * time.Hour)),
			expected:    false,
			description: "Ended series crawled 1 day ago should not refresh",
		},
		{
			name:        "Ended series - crawled 8 days ago",
			status:      "Ended",
			lastCrawled: timePtr(now.Add(-8 * 24 * time.Hour)),
			expected:    true,
			description: "Ended series crawled 8 days ago should refresh",
		},
		{
			name:        "Unknown status - never crawled",
			status:      "Unknown",
			lastCrawled: nil,
			expected:    true,
			description: "Unknown status with no crawl time should refresh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			show := &Show{
				Status:        tt.status,
				LastCrawledAt: tt.lastCrawled,
			}
			if got := show.ShouldRefresh(); got != tt.expected {
				t.Errorf("Show.ShouldRefresh() = %v, want %v (%s)", got, tt.expected, tt.description)
			}
		})
	}
}

func TestShow_TableName(t *testing.T) {
	show := Show{}
	if got := show.TableName(); got != "shows" {
		t.Errorf("Show.TableName() = %v, want %v", got, "shows")
	}
}

// Helper function
func timePtr(t time.Time) *time.Time {
	return &t
}
