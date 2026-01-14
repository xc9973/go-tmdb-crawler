package repositories

import (
	"github.com/xc9973/go-tmdb-crawler/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UploadedEpisodeRepository defines the interface for uploaded episode operations
type UploadedEpisodeRepository interface {
	MarkUploaded(episodeID uint) error
	UnmarkUploaded(episodeID uint) error
	IsUploaded(episodeID uint) (bool, error)
	GetByEpisodeID(episodeID uint) (*models.UploadedEpisode, error)
}

type uploadedEpisodeRepository struct {
	db *gorm.DB
}

// NewUploadedEpisodeRepository creates a new uploaded episode repository
func NewUploadedEpisodeRepository(db *gorm.DB) UploadedEpisodeRepository {
	return &uploadedEpisodeRepository{db: db}
}

// MarkUploaded marks an episode as uploaded (idempotent)
func (r *uploadedEpisodeRepository) MarkUploaded(episodeID uint) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "episode_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"uploaded", "updated_at"}),
	}).Create(&models.UploadedEpisode{
		EpisodeID: episodeID,
		Uploaded:  true,
	}).Error
}

// UnmarkUploaded removes the uploaded mark for an episode
func (r *uploadedEpisodeRepository) UnmarkUploaded(episodeID uint) error {
	return r.db.Where("episode_id = ?", episodeID).Delete(&models.UploadedEpisode{}).Error
}

// IsUploaded checks if an episode is marked as uploaded
func (r *uploadedEpisodeRepository) IsUploaded(episodeID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.UploadedEpisode{}).
		Where("episode_id = ? AND uploaded = ?", episodeID, true).
		Count(&count).Error
	return count > 0, err
}

// GetByEpisodeID retrieves the upload record for an episode
func (r *uploadedEpisodeRepository) GetByEpisodeID(episodeID uint) (*models.UploadedEpisode, error) {
	var ue models.UploadedEpisode
	err := r.db.Where("episode_id = ?", episodeID).First(&ue).Error
	if err != nil {
		return nil, err
	}
	return &ue, nil
}
