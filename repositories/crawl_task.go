package repositories

import (
	"time"

	"github.com/xc9973/go-tmdb-crawler/models"
	"gorm.io/gorm"
)

// CrawlTaskRepository defines data operations for crawl tasks
type CrawlTaskRepository interface {
	Create(task *models.CrawlTask) error
	Update(task *models.CrawlTask) error
	GetByID(id uint) (*models.CrawlTask, error)
	GetByStatus(status string, page, pageSize int) ([]*models.CrawlTask, int64, error)
	GetRecent(limit int) ([]*models.CrawlTask, error)
	GetRunning() ([]*models.CrawlTask, error)
	Delete(id uint) error
	DeleteOld(days int) error
	Count() (int64, error)
	CountByStatus(status string) (int64, error)
}

type crawlTaskRepository struct {
	db *gorm.DB
}

// NewCrawlTaskRepository creates a new crawl task repository instance
func NewCrawlTaskRepository(db *gorm.DB) CrawlTaskRepository {
	return &crawlTaskRepository{db: db}
}

// Create creates a new crawl task
func (r *crawlTaskRepository) Create(task *models.CrawlTask) error {
	return r.db.Create(task).Error
}

// Update updates a crawl task
func (r *crawlTaskRepository) Update(task *models.CrawlTask) error {
	return r.db.Save(task).Error
}

// GetByID retrieves a crawl task by ID
func (r *crawlTaskRepository) GetByID(id uint) (*models.CrawlTask, error) {
	var task models.CrawlTask
	if err := r.db.First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// GetByStatus retrieves crawl tasks by status with pagination
func (r *crawlTaskRepository) GetByStatus(status string, page, pageSize int) ([]*models.CrawlTask, int64, error) {
	var tasks []*models.CrawlTask
	var total int64

	query := r.db.Model(&models.CrawlTask{}).Where("status = ?", status)

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated data
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&tasks).Error

	return tasks, total, err
}

// GetRecent retrieves recent crawl tasks
func (r *crawlTaskRepository) GetRecent(limit int) ([]*models.CrawlTask, error) {
	var tasks []*models.CrawlTask
	err := r.db.Order("created_at DESC").
		Limit(limit).
		Find(&tasks).Error
	return tasks, err
}

// GetRunning retrieves all currently running tasks
func (r *crawlTaskRepository) GetRunning() ([]*models.CrawlTask, error) {
	var tasks []*models.CrawlTask
	err := r.db.Where("status = ?", "running").
		Order("started_at ASC").
		Find(&tasks).Error
	return tasks, err
}

// Delete deletes a crawl task by ID
func (r *crawlTaskRepository) Delete(id uint) error {
	return r.db.Delete(&models.CrawlTask{}, id).Error
}

// DeleteOld deletes crawl tasks older than specified days
func (r *crawlTaskRepository) DeleteOld(days int) error {
	cutoffDate := time.Now().AddDate(0, 0, -days)
	return r.db.Where("created_at < ?", cutoffDate).Delete(&models.CrawlTask{}).Error
}

// Count returns the total number of crawl tasks
func (r *crawlTaskRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.CrawlTask{}).Count(&count).Error
	return count, err
}

// CountByStatus returns the count of tasks by status
func (r *crawlTaskRepository) CountByStatus(status string) (int64, error) {
	var count int64
	err := r.db.Model(&models.CrawlTask{}).
		Where("status = ?", status).
		Count(&count).Error
	return count, err
}
