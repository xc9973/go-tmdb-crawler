package models

import (
	"testing"
	"time"
)

func TestEpisode_GetEpisodeCode(t *testing.T) {
	tests := []struct {
		name          string
		seasonNumber  int
		episodeNumber int
		expected      string
	}{
		{
			name:          "Season 1 Episode 1",
			seasonNumber:  1,
			episodeNumber: 1,
			expected:      "S01E01",
		},
		{
			name:          "Season 10 Episode 15",
			seasonNumber:  10,
			episodeNumber: 15,
			expected:      "S10E15",
		},
		{
			name:          "Season 0 Episode 1",
			seasonNumber:  0,
			episodeNumber: 1,
			expected:      "S00E01",
		},
		{
			name:          "Season 1 Episode 0",
			seasonNumber:  1,
			episodeNumber: 0,
			expected:      "S01E00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ep := &Episode{
				SeasonNumber:  tt.seasonNumber,
				EpisodeNumber: tt.episodeNumber,
			}
			if got := ep.GetEpisodeCode(); got != tt.expected {
				t.Errorf("Episode.GetEpisodeCode() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEpisode_IsAired(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		airDate  *time.Time
		expected bool
	}{
		{
			name:     "Aired in the past",
			airDate:  timePtr(now.Add(-24 * time.Hour)),
			expected: true,
		},
		{
			name:     "Aired just now",
			airDate:  timePtr(now.Add(-1 * time.Second)),
			expected: true,
		},
		{
			name:     "Will air in the future",
			airDate:  timePtr(now.Add(24 * time.Hour)),
			expected: false,
		},
		{
			name:     "No air date",
			airDate:  nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ep := &Episode{AirDate: tt.airDate}
			if got := ep.IsAired(); got != tt.expected {
				t.Errorf("Episode.IsAired() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEpisode_IsAiredInTimezone(t *testing.T) {
	now := time.Now()
	ny, _ := time.LoadLocation("America/New_York")
	tokyo, _ := time.LoadLocation("Asia/Tokyo")

	tests := []struct {
		name     string
		airDate  *time.Time
		location *time.Location
		expected bool
	}{
		{
			name:     "Aired in New York timezone",
			airDate:  timePtr(now.Add(-24 * time.Hour)),
			location: ny,
			expected: true,
		},
		{
			name:     "Will air in Tokyo timezone",
			airDate:  timePtr(now.Add(24 * time.Hour)),
			location: tokyo,
			expected: false,
		},
		{
			name:     "No air date",
			airDate:  nil,
			location: ny,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ep := &Episode{AirDate: tt.airDate}
			if got := ep.IsAiredInTimezone(tt.location); got != tt.expected {
				t.Errorf("Episode.IsAiredInTimezone() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEpisode_IsFuture(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		airDate  *time.Time
		expected bool
	}{
		{
			name:     "Will air in the future",
			airDate:  timePtr(now.Add(24 * time.Hour)),
			expected: true,
		},
		{
			name:     "Will air in 1 second",
			airDate:  timePtr(now.Add(1 * time.Second)),
			expected: true,
		},
		{
			name:     "Aired in the past",
			airDate:  timePtr(now.Add(-24 * time.Hour)),
			expected: false,
		},
		{
			name:     "No air date",
			airDate:  nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ep := &Episode{AirDate: tt.airDate}
			if got := ep.IsFuture(); got != tt.expected {
				t.Errorf("Episode.IsFuture() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEpisode_IsFutureInTimezone(t *testing.T) {
	now := time.Now()
	ny, _ := time.LoadLocation("America/New_York")

	tests := []struct {
		name     string
		airDate  *time.Time
		location *time.Location
		expected bool
	}{
		{
			name:     "Will air in New York timezone",
			airDate:  timePtr(now.Add(24 * time.Hour)),
			location: ny,
			expected: true,
		},
		{
			name:     "Aired in New York timezone",
			airDate:  timePtr(now.Add(-24 * time.Hour)),
			location: ny,
			expected: false,
		},
		{
			name:     "No air date",
			airDate:  nil,
			location: ny,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ep := &Episode{AirDate: tt.airDate}
			if got := ep.IsFutureInTimezone(tt.location); got != tt.expected {
				t.Errorf("Episode.IsFutureInTimezone() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEpisode_Validate(t *testing.T) {
	tests := []struct {
		name    string
		ep      *Episode
		wantErr bool
	}{
		{
			name: "Valid episode",
			ep: &Episode{
				ShowID:        1,
				SeasonNumber:  1,
				EpisodeNumber: 1,
			},
			wantErr: false,
		},
		{
			name: "Empty ShowID",
			ep: &Episode{
				ShowID:        0,
				SeasonNumber:  1,
				EpisodeNumber: 1,
			},
			wantErr: true,
		},
		{
			name: "Negative season number",
			ep: &Episode{
				ShowID:        1,
				SeasonNumber:  -1,
				EpisodeNumber: 1,
			},
			wantErr: true,
		},
		{
			name: "Negative episode number",
			ep: &Episode{
				ShowID:        1,
				SeasonNumber:  1,
				EpisodeNumber: -1,
			},
			wantErr: true,
		},
		{
			name: "Zero season and episode numbers",
			ep: &Episode{
				ShowID:        1,
				SeasonNumber:  0,
				EpisodeNumber: 0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.ep.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Episode.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEpisode_GetAirDateFormatted(t *testing.T) {
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		airDate  *time.Time
		expected string
	}{
		{
			name:     "Valid air date",
			airDate:  timePtr(date),
			expected: "2024-01-15",
		},
		{
			name:     "No air date",
			airDate:  nil,
			expected: "TBD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ep := &Episode{AirDate: tt.airDate}
			if got := ep.GetAirDateFormatted(); got != tt.expected {
				t.Errorf("Episode.GetAirDateFormatted() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEpisode_GetAirDateFormattedInTimezone(t *testing.T) {
	ny, _ := time.LoadLocation("America/New_York")
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		airDate  *time.Time
		location *time.Location
		expected string
	}{
		{
			name:     "Valid air date in New York",
			airDate:  timePtr(date),
			location: ny,
			expected: "2024-01-14", // UTC midnight is previous day in NY
		},
		{
			name:     "No air date",
			airDate:  nil,
			location: ny,
			expected: "TBD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ep := &Episode{AirDate: tt.airDate}
			if got := ep.GetAirDateFormattedInTimezone(tt.location); got != tt.expected {
				t.Errorf("Episode.GetAirDateFormattedInTimezone() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEpisode_TableName(t *testing.T) {
	ep := Episode{}
	if got := ep.TableName(); got != "episodes" {
		t.Errorf("Episode.TableName() = %v, want %v", got, "episodes")
	}
}
