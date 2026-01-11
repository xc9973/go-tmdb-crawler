package repositories

import (
	"time"

	"github.com/yourusername/go-tmdb-crawler/models"
	"gorm.io/gorm"
)

// EpisodeRepository defines the interface for episode data operations
type EpisodeRepository interface {
	Create(episode *models.Episode) error
	CreateBatch(episodes []*models.Episode) error
	GetByID(id uint) (*models.Episode, error)
	GetByShowID(showID uint) ([]*models.Episode, error)
	GetBySeason(showID uint, seasonNumber int) ([]*models.Episode, error)
	GetByDateRange(startDate, endDate time.Time) ([]*models.Episode, error)
	GetTodayUpdates() ([]*models.Episode, error)
	Update(episode *models.Episode) error
	Delete(id uint) error
	DeleteByShowID(showID uint) error
	CountByShowID(showID uint) (int64, error)
	Count() (int64, error)
}

type episodeRepository struct {
	db *gorm.DB
}

// NewEpisodeRepository creates a new episode repository instance
func NewEpisodeRepository(db *gorm.DB) EpisodeRepository {
	return &episodeRepository{db: db}
}

// Create creates a new episode
func (r *episodeRepository) Create(episode *models.Episode) error {
	return r.db.Create(episode).Error
}

// CreateBatch creates multiple episodes in a single transaction
func (r *episodeRepository) CreateBatch(episodes []*models.Episode) error {
	if len(episodes) == 0 {
		return nil
	}
	return r.db.CreateInBatches(episodes, 100).Error
}

// GetByID retrieves an episode by ID
func (r *episodeRepository) GetByID(id uint) (*models.Episode, error) {
	var episode models.Episode
	err := r.db.Preload("Show").First(&episode, id).Error
	if err != nil {
		return nil, err
	}
	return &episode, nil
}

// GetByShowID retrieves all episodes for a show
func (r *episodeRepository) GetByShowID(showID uint) ([]*models.Episode, error) {
	var episodes []*models.Episode
	err := r.db.Where("show_id = ?", showID).
		Order("season_number ASC, episode_number ASC").
		Find(&episodes).Error
	return episodes, err
}

// GetBySeason retrieves all episodes for a specific season
func (r *episodeRepository) GetBySeason(showID uint, seasonNumber int) ([]*models.Episode, error) {
	var episodes []*models.Episode
	err := r.db.Where("show_id = ? AND season_number = ?", showID, seasonNumber).
		Order("episode_number ASC").
		Find(&episodes).Error
	return episodes, err
}

// GetByDateRange retrieves episodes within a date range
func (r *episodeRepository) GetByDateRange(startDate, endDate time.Time) ([]*models.Episode, error) {
	var episodes []*models.Episode
	err := r.db.Where("air_date >= ? AND air_date <= ?", startDate, endDate).
		Preload("Show").
		Order("air_date ASC").
		Find(&episodes).Error
	return episodes, err
}

// GetTodayUpdates retrieves episodes airing today
func (r *episodeRepository) GetTodayUpdates() ([]*models.Episode, error) {
	var episodes []*models.Episode
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	err := r.db.Where("air_date >= ? AND air_date < ?", today, tomorrow).
		Preload("Show").
		Order("air_date ASC").
		Find(&episodes).Error
	return episodes, err
}

// Update updates an episode
func (r *episodeRepository) Update(episode *models.Episode) error {
	return r.db.Save(episode).Error
}

// Delete deletes an episode by ID
func (r *episodeRepository) Delete(id uint) error {
	return r.db.Delete(&models.Episode{}, id).Error
}

// DeleteByShowID deletes all episodes for a show
func (r *episodeRepository) DeleteByShowID(showID uint) error {
	return r.db.Where("show_id = ?", showID).Delete(&models.Episode{}).Error
}

// CountByShowID returns the total number of episodes for a show
func (r *episodeRepository) CountByShowID(showID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Episode{}).
		Where("show_id = ?", showID).
		Count(&count).Error
	return count, err
}

// Count returns the total number of episodes
func (r *episodeRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.Episode{}).Count(&count).Error
	return count, err
}
