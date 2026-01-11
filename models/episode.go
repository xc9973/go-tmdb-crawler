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

// TableName specifies the table name for Episode model
func (Episode) TableName() string {
	return "episodes"
}

// GetEpisodeCode returns the episode code in format S01E01
func (e *Episode) GetEpisodeCode() string {
	return fmt.Sprintf("S%02dE%02d", e.SeasonNumber, e.EpisodeNumber)
}

// IsAired checks if the episode has already aired
func (e *Episode) IsAired() bool {
	if e.AirDate == nil {
		return false
	}
	return e.AirDate.Before(time.Now())
}

// IsFuture checks if the episode will air in the future
func (e *Episode) IsFuture() bool {
	if e.AirDate == nil {
		return false
	}
	return e.AirDate.After(time.Now())
}

// BeforeCreate hook
func (e *Episode) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	e.CreatedAt = now
	e.UpdatedAt = now
	return nil
}
