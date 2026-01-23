package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Show represents a TV show from TMDB
type Show struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	TmdbID       int        `gorm:"uniqueIndex:idx_tmdb_id;not null" json:"tmdb_id"`
	Name         string     `gorm:"size:255;not null;index:idx_name" json:"name"`
	OriginalName string     `gorm:"size:255" json:"original_name"`
	Status       string     `gorm:"size:50;index:idx_status" json:"status"`
	Type         string     `gorm:"size:50" json:"type"`
	Language     string     `gorm:"size:10" json:"language"`
	FirstAirDate *time.Time `gorm:"index:idx_first_air_date" json:"first_air_date"`
	Overview     string     `gorm:"type:text" json:"overview"`
	PosterPath   string     `gorm:"size:512" json:"poster_path"`
	BackdropPath string     `gorm:"size:512" json:"backdrop_path"`
	Genres       string     `gorm:"size:255" json:"genres"`
	Popularity   float64    `gorm:"type:decimal(5,2);default:0.0" json:"popularity"`
	VoteAverage  float32    `gorm:"type:decimal(3,1);default:0.0" json:"vote_average"`
	VoteCount    int        `gorm:"default:0" json:"vote_count"`

	// Local fields
	LastSeasonNumber int        `gorm:"default:0" json:"last_season_number"`
	LastEpisodeCount int        `gorm:"default:0" json:"last_episode_count"`
	NextAirDate      *time.Time `gorm:"index:idx_next_air_date" json:"next_air_date"`
	CustomStatus     string     `gorm:"size:50" json:"custom_status"`
	Notes            string     `gorm:"type:text" json:"notes"`

	// Correction fields
	RefreshThreshold      int        `gorm:"default:0" json:"refresh_threshold"`
	StaleDetectedAt      *time.Time `gorm:"index:idx_stale_detected_at" json:"stale_detected_at"`
	LastCorrectionResult string     `gorm:"size:50" json:"last_correction_result"`

	// Timestamps
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	LastCrawledAt *time.Time `gorm:"index:idx_last_crawled" json:"last_crawled_at"`

	// Relationships
	Episodes []Episode `gorm:"foreignKey:ShowID;constraint:OnDelete:CASCADE" json:"episodes,omitempty"`
}

// TableName specifies the table name for Show model
func (Show) TableName() string {
	return "shows"
}

// IsReturning checks if the show is still returning/airing
func (s *Show) IsReturning() bool {
	return s.Status == "Returning Series"
}

// IsEnded checks if the show has ended
func (s *Show) IsEnded() bool {
	return s.Status == "Ended"
}

// GetDisplayStatus returns the display status (custom or original)
func (s *Show) GetDisplayStatus() string {
	if s.CustomStatus != "" {
		return s.CustomStatus
	}
	return s.Status
}

// GetDisplayType returns the display type
func (s *Show) GetDisplayType() string {
	if s.Type != "" {
		return s.Type
	}
	return "Unknown"
}

// BeforeCreate hook
func (s *Show) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	s.CreatedAt = now
	s.UpdatedAt = now
	return s.Validate()
}

// BeforeUpdate hook
func (s *Show) BeforeUpdate(tx *gorm.DB) error {
	s.UpdatedAt = time.Now()
	return s.Validate()
}

// Validate validates the show data
func (s *Show) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("show name cannot be empty")
	}
	if s.TmdbID <= 0 {
		return fmt.Errorf("invalid TMDB ID")
	}
	return nil
}

// IsExpired checks if the show data needs to be refreshed (older than 24 hours)
func (s *Show) IsExpired() bool {
	if s.LastCrawledAt == nil {
		return true
	}
	return time.Since(*s.LastCrawledAt) > 24*time.Hour
}

// ShouldRefresh checks if the show should be refreshed based on status and last crawl time
func (s *Show) ShouldRefresh() bool {
	// Always refresh returning series
	if s.IsReturning() {
		return s.IsExpired()
	}
	// Refresh ended shows if data is older than 7 days
	if s.IsEnded() && s.LastCrawledAt != nil {
		return time.Since(*s.LastCrawledAt) > 7*24*time.Hour
	}
	return s.IsExpired()
}
