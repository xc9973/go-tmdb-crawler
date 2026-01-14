package repositories

import (
	"time"

	"github.com/xc9973/go-tmdb-crawler/models"
	"github.com/xc9973/go-tmdb-crawler/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	GetTodayUpdatesWithUploadStatus() ([]map[string]interface{}, error)
	Update(episode *models.Episode) error
	Delete(id uint) error
	DeleteByShowID(showID uint) error
	CountByShowID(showID uint) (int64, error)
	Count() (int64, error)
	SetTimezoneHelper(tzHelper *utils.TimezoneHelper)
}

type episodeRepository struct {
	db             *gorm.DB
	timezoneHelper *utils.TimezoneHelper
}

// NewEpisodeRepository creates a new episode repository instance
func NewEpisodeRepository(db *gorm.DB) EpisodeRepository {
	// Default to UTC if no timezone specified
	location, _ := time.LoadLocation("UTC")
	return &episodeRepository{
		db:             db,
		timezoneHelper: utils.NewTimezoneHelper(location),
	}
}

// SetTimezoneHelper sets the timezone helper for date operations
// This should be called during application initialization
func (r *episodeRepository) SetTimezoneHelper(tzHelper *utils.TimezoneHelper) {
	r.timezoneHelper = tzHelper
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
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "show_id"},
			{Name: "season_number"},
			{Name: "episode_number"},
		},
		DoNothing: true,
	}).CreateInBatches(episodes, 100).Error
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
// The range is inclusive: [startDate, endDate]
// Dates are interpreted in the configured timezone
func (r *episodeRepository) GetByDateRange(startDate, endDate time.Time) ([]*models.Episode, error) {
	var episodes []*models.Episode

	// Use timezone-aware date boundaries
	start, end := r.timezoneHelper.DateRange(startDate, endDate)

	err := r.db.Where("air_date >= ? AND air_date <= ?", start, end).
		Preload("Show").
		Order("air_date ASC").
		Find(&episodes).Error
	return episodes, err
}

// GetTodayUpdates retrieves episodes airing today
// Today is determined based on the configured timezone
// The range is [startOfDay, endOfDay) - start inclusive, end exclusive
func (r *episodeRepository) GetTodayUpdates() ([]*models.Episode, error) {
	var episodes []*models.Episode

	// Use timezone-aware today boundaries
	start, end := r.timezoneHelper.TodayRange()

	err := r.db.Where("air_date >= ? AND air_date < ?", start, end).
		Preload("Show").
		Order("air_date ASC").
		Find(&episodes).Error
	return episodes, err
}

// GetTodayUpdatesWithUploadStatus retrieves episodes airing today with upload status
// 返回结构与前端 today.js 期望的格式匹配
func (r *episodeRepository) GetTodayUpdatesWithUploadStatus() ([]map[string]interface{}, error) {
	start, end := r.timezoneHelper.TodayRange()

	type Result struct {
		ID            uint
		SeasonNumber  int
		EpisodeNumber int
		Name          string
		AirDate       *time.Time
		StillPath     string
		VoteAverage   float32
		ShowID        uint
		ShowName      string
		PosterPath    string
		ShowStatus    string
		Uploaded      bool
	}

	var results []Result
	err := r.db.Raw(`
        SELECT
            e.id,
            e.season_number,
            e.episode_number,
            e.name,
            e.air_date,
            e.still_path,
            e.vote_average,
            e.show_id,
            s.name as show_name,
            s.poster_path,
            s.status as show_status,
            COALESCE(ue.uploaded, 0) as uploaded
        FROM episodes e
        INNER JOIN shows s ON e.show_id = s.id
        LEFT JOIN uploaded_episodes ue ON e.id = ue.episode_id
        WHERE e.air_date >= ? AND e.air_date < ?
        ORDER BY e.air_date ASC
    `, start, end).Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// 转换为 map 格式以匹配前端期望
	episodes := make([]map[string]interface{}, len(results))
	for i, r := range results {
		episodes[i] = map[string]interface{}{
			"id":             r.ID,
			"season_number":  r.SeasonNumber,
			"episode_number": r.EpisodeNumber,
			"name":           r.Name,
			"air_date":       r.AirDate,
			"still_path":     r.StillPath,
			"vote_average":   r.VoteAverage,
			"show_id":        r.ShowID,
			"show_name":      r.ShowName,
			"poster_path":    r.PosterPath,
			"show_status":    r.ShowStatus,
			"uploaded":       r.Uploaded,
		}
	}

	return episodes, nil
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
