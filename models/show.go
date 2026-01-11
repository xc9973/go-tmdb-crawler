package models

import (
	"time"

	"gorm.io/gorm"
)

// Show represents a TV show from TMDB
type Show struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	TmdbID       int        `gorm:"uniqueIndex;not null" json:"tmdb_id"`
	Name         string     `gorm:"size:255;not null" json:"name"`
	OriginalName string     `gorm:"size:255" json:"original_name"`
	Status       string     `gorm:"size:50" json:"status"`
	Type         string     `gorm:"size:50" json:"type"`
	Language     string     `gorm:"size:10" json:"language"`
	FirstAirDate *time.Time `json:"first_air_date"`
	Overview     string     `gorm:"type:text" json:"overview"`
	PosterPath   string     `gorm:"size:512" json:"poster_path"`
	BackdropPath string     `gorm:"size:512" json:"backdrop_path"`
	Genres       string     `gorm:"size:255" json:"genres"`
	Popularity   float64    `gorm:"type:decimal(5,2)" json:"popularity"`
	VoteAverage  float32    `gorm:"type:decimal(3,1)" json:"vote_average"`
	VoteCount    int        `json:"vote_count"`

	// Local fields
	LastSeasonNumber int        `json:"last_season_number"`
	LastEpisodeCount int        `json:"last_episode_count"`
	NextAirDate      *time.Time `json:"next_air_date"`
	CustomStatus     string     `gorm:"size:50" json:"custom_status"`
	Notes            string     `gorm:"type:text" json:"notes"`

	// Timestamps
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	LastCrawledAt *time.Time `json:"last_crawled_at"`

	// Relationships
	Episodes []Episode `gorm:"foreignKey:ShowID" json:"episodes,omitempty"`
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
	return nil
}
