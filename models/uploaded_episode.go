package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// UploadedEpisode tracks which episodes have been uploaded to NAS
type UploadedEpisode struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	EpisodeID uint      `gorm:"not null;uniqueIndex" json:"episode_id"`
	Uploaded  bool      `gorm:"not null;default:true" json:"uploaded"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (UploadedEpisode) TableName() string {
	return "uploaded_episodes"
}

// BeforeCreate hook sets timestamps
func (ue *UploadedEpisode) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	ue.CreatedAt = now
	ue.UpdatedAt = now
	return ue.Validate()
}

// BeforeUpdate hook sets updated_at
func (ue *UploadedEpisode) BeforeUpdate(tx *gorm.DB) error {
	ue.UpdatedAt = time.Now()
	return ue.Validate()
}

// Validate validates the uploaded episode data
func (ue *UploadedEpisode) Validate() error {
	if ue.EpisodeID == 0 {
		return errors.New("episode ID cannot be empty")
	}
	return nil
}
