package repositories

import (
	"github.com/xc9973/go-tmdb-crawler/models"
	"gorm.io/gorm"
)

// CrawlTaskRepository defines data operations for crawl tasks
// NOTE: keep interface minimal to avoid over-coupling with scheduler.
type CrawlTaskRepository interface {
	Create(task *models.CrawlTask) error
	Update(task *models.CrawlTask) error
	GetByID(id uint) (*models.CrawlTask, error)
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
