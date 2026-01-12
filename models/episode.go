package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Episode represents a single episode of a TV show
type Episode struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	ShowID        uint       `gorm:"not null;index" json:"show_id"`
	SeasonNumber  int        `gorm:"not null" json:"season_number"`
	EpisodeNumber int        `gorm:"not null" json:"episode_number"`
	Name          string     `gorm:"size:255" json:"name"`
	Overview      string     `gorm:"type:text" json:"overview"`
	AirDate       *time.Time `gorm:"index" json:"air_date"`
	StillPath     string     `gorm:"size:512" json:"still_path"`
	Runtime       int        `json:"runtime"` // in minutes
	VoteAverage   float32    `gorm:"type:decimal(3,1)" json:"vote_average"`
	VoteCount     int        `json:"vote_count"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Show *Show `gorm:"foreignKey:ShowID" json:"show,omitempty"`
}

// EpisodeUniqueIndex defines the unique constraint for episodes
// This ensures no duplicate episodes for the same show, season, and episode number
// Note: GORM AutoMigrate will not create this constraint automatically.
// For SQLite, you need to create the unique index manually via migration SQL.
// For PostgreSQL, GORM may create it but it's recommended to use explicit migrations.
// See migrations/001_init_schema.sql for the constraint definition.
func (Episode) TableName() string {
	return "episodes"
}

// GetEpisodeCode returns the episode code in format S01E01
func (e *Episode) GetEpisodeCode() string {
	return fmt.Sprintf("S%02dE%02d", e.SeasonNumber, e.EpisodeNumber)
}

// IsAired checks if the episode has already aired
// Uses current time in the system's local timezone for comparison
func (e *Episode) IsAired() bool {
	if e.AirDate == nil {
		return false
	}
	return e.AirDate.Before(time.Now())
}

// IsAiredInTimezone checks if the episode has already aired in the specified timezone
func (e *Episode) IsAiredInTimezone(loc *time.Location) bool {
	if e.AirDate == nil {
		return false
	}
	return e.AirDate.Before(time.Now().In(loc))
}

// IsFuture checks if the episode will air in the future
// Uses current time in the system's local timezone for comparison
func (e *Episode) IsFuture() bool {
	if e.AirDate == nil {
		return false
	}
	return e.AirDate.After(time.Now())
}

// IsFutureInTimezone checks if the episode will air in the future in the specified timezone
func (e *Episode) IsFutureInTimezone(loc *time.Location) bool {
	if e.AirDate == nil {
		return false
	}
	return e.AirDate.After(time.Now().In(loc))
}

// BeforeCreate hook
func (e *Episode) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	e.CreatedAt = now
	e.UpdatedAt = now
	return nil
}
