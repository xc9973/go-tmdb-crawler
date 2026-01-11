package repositories

import (
	"time"

	"github.com/yourusername/go-tmdb-crawler/models"
	"gorm.io/gorm"
)

// CrawlLogRepository defines the interface for crawl log data operations
type CrawlLogRepository interface {
	Create(log *models.CrawlLog) error
	GetByID(id uint) (*models.CrawlLog, error)
	GetByShowID(showID uint, limit int) ([]*models.CrawlLog, error)
	GetRecent(limit int) ([]*models.CrawlLog, error)
	GetByStatus(status string, page, pageSize int) ([]*models.CrawlLog, int64, error)
	GetByDateRange(startDate, endDate time.Time) ([]*models.CrawlLog, error)
	Delete(id uint) error
	DeleteOld(days int) error
	Count() (int64, error)
	CountByStatus(status string) (int64, error)
}

type crawlLogRepository struct {
	db *gorm.DB
}

// NewCrawlLogRepository creates a new crawl log repository instance
func NewCrawlLogRepository(db *gorm.DB) CrawlLogRepository {
	return &crawlLogRepository{db: db}
}

// Create creates a new crawl log
func (r *crawlLogRepository) Create(log *models.CrawlLog) error {
	return r.db.Create(log).Error
}

// GetByID retrieves a crawl log by ID
func (r *crawlLogRepository) GetByID(id uint) (*models.CrawlLog, error) {
	var log models.CrawlLog
	err := r.db.Preload("Show").First(&log, id).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// GetByShowID retrieves crawl logs for a specific show
func (r *crawlLogRepository) GetByShowID(showID uint, limit int) ([]*models.CrawlLog, error) {
	var logs []*models.CrawlLog
	query := r.db.Where("show_id = ?", showID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&logs).Error
	return logs, err
}

// GetRecent retrieves recent crawl logs
func (r *crawlLogRepository) GetRecent(limit int) ([]*models.CrawlLog, error) {
	var logs []*models.CrawlLog
	err := r.db.Preload("Show").
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// GetByStatus retrieves crawl logs by status with pagination
func (r *crawlLogRepository) GetByStatus(status string, page, pageSize int) ([]*models.CrawlLog, int64, error) {
	var logs []*models.CrawlLog
	var total int64

	query := r.db.Model(&models.CrawlLog{}).Where("status = ?", status)

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated data
	offset := (page - 1) * pageSize
	err := query.Preload("Show").
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&logs).Error

	return logs, total, err
}

// GetByDateRange retrieves crawl logs within a date range
func (r *crawlLogRepository) GetByDateRange(startDate, endDate time.Time) ([]*models.CrawlLog, error) {
	var logs []*models.CrawlLog
	err := r.db.Where("created_at >= ? AND created_at <= ?", startDate, endDate).
		Preload("Show").
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}

// Delete deletes a crawl log by ID
func (r *crawlLogRepository) Delete(id uint) error {
	return r.db.Delete(&models.CrawlLog{}, id).Error
}

// DeleteOld deletes crawl logs older than specified days
func (r *crawlLogRepository) DeleteOld(days int) error {
	cutoffDate := time.Now().AddDate(0, 0, -days)
	return r.db.Where("created_at < ?", cutoffDate).Delete(&models.CrawlLog{}).Error
}

// Count returns the total number of crawl logs
func (r *crawlLogRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.CrawlLog{}).Count(&count).Error
	return count, err
}

// CountByStatus returns the count of logs by status
func (r *crawlLogRepository) CountByStatus(status string) (int64, error) {
	var count int64
	err := r.db.Model(&models.CrawlLog{}).
		Where("status = ?", status).
		Count(&count).Error
	return count, err
}
