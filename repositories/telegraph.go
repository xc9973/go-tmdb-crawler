package repositories

import (
	"time"

	"github.com/xc9973/go-tmdb-crawler/models"
	"gorm.io/gorm"
)

// TelegraphPostRepository defines the interface for telegraph post data operations
type TelegraphPostRepository interface {
	Create(post *models.TelegraphPost) error
	GetByID(id uint) (*models.TelegraphPost, error)
	GetByPath(path string) (*models.TelegraphPost, error)
	GetByContentHash(hash string) (*models.TelegraphPost, error)
	GetRecent(limit int) ([]*models.TelegraphPost, error)
	GetToday() (*models.TelegraphPost, error)
	GetByDateRange(startDate, endDate time.Time) ([]*models.TelegraphPost, error)
	Update(post *models.TelegraphPost) error
	Delete(id uint) error
	DeleteOld(days int) error
	Count() (int64, error)
	CountToday() (int64, error)
}

type telegraphPostRepository struct {
	db *gorm.DB
}

// NewTelegraphPostRepository creates a new telegraph post repository instance
func NewTelegraphPostRepository(db *gorm.DB) TelegraphPostRepository {
	return &telegraphPostRepository{db: db}
}

// Create creates a new telegraph post
func (r *telegraphPostRepository) Create(post *models.TelegraphPost) error {
	return r.db.Create(post).Error
}

// GetByID retrieves a telegraph post by ID
func (r *telegraphPostRepository) GetByID(id uint) (*models.TelegraphPost, error) {
	var post models.TelegraphPost
	err := r.db.First(&post, id).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

// GetByPath retrieves a telegraph post by telegraph path
func (r *telegraphPostRepository) GetByPath(path string) (*models.TelegraphPost, error) {
	var post models.TelegraphPost
	err := r.db.Where("telegraph_path = ?", path).First(&post).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

// GetByContentHash retrieves a telegraph post by content hash
func (r *telegraphPostRepository) GetByContentHash(hash string) (*models.TelegraphPost, error) {
	var post models.TelegraphPost
	err := r.db.Where("content_hash = ?", hash).
		Order("created_at DESC").
		First(&post).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

// GetRecent retrieves recent telegraph posts
func (r *telegraphPostRepository) GetRecent(limit int) ([]*models.TelegraphPost, error) {
	var posts []*models.TelegraphPost
	err := r.db.Order("created_at DESC").
		Limit(limit).
		Find(&posts).Error
	return posts, err
}

// GetToday retrieves the most recent post created today
func (r *telegraphPostRepository) GetToday() (*models.TelegraphPost, error) {
	var post models.TelegraphPost
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	err := r.db.Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay).
		Order("created_at DESC").
		First(&post).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

// GetByDateRange retrieves telegraph posts within a date range
func (r *telegraphPostRepository) GetByDateRange(startDate, endDate time.Time) ([]*models.TelegraphPost, error) {
	var posts []*models.TelegraphPost
	err := r.db.Where("created_at >= ? AND created_at <= ?", startDate, endDate).
		Order("created_at DESC").
		Find(&posts).Error
	return posts, err
}

// Update updates a telegraph post
func (r *telegraphPostRepository) Update(post *models.TelegraphPost) error {
	return r.db.Save(post).Error
}

// Delete deletes a telegraph post by ID
func (r *telegraphPostRepository) Delete(id uint) error {
	return r.db.Delete(&models.TelegraphPost{}, id).Error
}

// DeleteOld deletes telegraph posts older than specified days
func (r *telegraphPostRepository) DeleteOld(days int) error {
	cutoffDate := time.Now().AddDate(0, 0, -days)
	return r.db.Where("created_at < ?", cutoffDate).Delete(&models.TelegraphPost{}).Error
}

// Count returns the total number of telegraph posts
func (r *telegraphPostRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.TelegraphPost{}).Count(&count).Error
	return count, err
}

// CountToday returns the number of posts created today
func (r *telegraphPostRepository) CountToday() (int64, error) {
	var count int64
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	err := r.db.Model(&models.TelegraphPost{}).
		Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay).
		Count(&count).Error
	return count, err
}
