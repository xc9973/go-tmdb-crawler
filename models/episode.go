package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Episode represents a single episode of a TV show
type Episode struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	ShowID        uint       `gorm:"not null;index:idx_show_id,priority:1" json:"show_id"`
	SeasonNumber  int        `gorm:"not null;index:idx_show_id,priority:2;index:idx_season_number" json:"season_number"`
	EpisodeNumber int        `gorm:"not null;index:idx_show_id,priority:3" json:"episode_number"`
	Name          string     `gorm:"size:255" json:"name"`
	Overview      string     `gorm:"type:text" json:"overview"`
	AirDate       *time.Time `gorm:"index:idx_air_date" json:"air_date"`
	StillPath     string     `gorm:"size:512" json:"still_path"`
	Runtime       int        `gorm:"default:0" json:"runtime"` // in minutes
	VoteAverage   float32    `gorm:"type:decimal(3,1);default:0.0" json:"vote_average"`
	VoteCount     int        `gorm:"default:0" json:"vote_count"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Show *Show `gorm:"foreignKey:ShowID;constraint:OnDelete:CASCADE" json:"show,omitempty"`
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
	return e.Validate()
}

// BeforeUpdate hook
func (e *Episode) BeforeUpdate(tx *gorm.DB) error {
	e.UpdatedAt = time.Now()
	return e.Validate()
}

// Validate validates the episode data
func (e *Episode) Validate() error {
	if e.ShowID == 0 {
		return fmt.Errorf("show ID cannot be empty")
	}
	if e.SeasonNumber < 0 {
		return fmt.Errorf("season number cannot be negative")
	}
	if e.EpisodeNumber < 0 {
		return fmt.Errorf("episode number cannot be negative")
	}
	return nil
}

// GetAirDateFormatted returns the air date in a formatted string
func (e *Episode) GetAirDateFormatted() string {
	if e.AirDate == nil {
		return "TBD"
	}
	return e.AirDate.Format("2006-01-02")
}

// GetAirDateFormattedInTimezone returns the air date in a formatted string for the given timezone
func (e *Episode) GetAirDateFormattedInTimezone(loc *time.Location) string {
	if e.AirDate == nil {
		return "TBD"
	}
	return e.AirDate.In(loc).Format("2006-01-02")
}
