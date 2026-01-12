package repositories

import (
	"strings"

	"github.com/xc9973/go-tmdb-crawler/models"
	"gorm.io/gorm"
)

// ShowRepository defines the interface for show data operations
type ShowRepository interface {
	Create(show *models.Show) error
	GetByID(id uint) (*models.Show, error)
	GetByTmdbID(tmdbID int) (*models.Show, error)
	List(page, pageSize int) ([]*models.Show, int64, error)
	ListByStatus(status string, page, pageSize int) ([]*models.Show, int64, error)
	ListFiltered(status, search string, page, pageSize int) ([]*models.Show, int64, error)
	ListAll() ([]*models.Show, error)
	ListReturning() ([]*models.Show, error)
	Update(show *models.Show) error
	Delete(id uint) error
	Count() (int64, error)
	Search(query string, page, pageSize int) ([]*models.Show, int64, error)
}

type showRepository struct {
	db *gorm.DB
}

// NewShowRepository creates a new show repository instance
func NewShowRepository(db *gorm.DB) ShowRepository {
	return &showRepository{db: db}
}

// Create creates a new show
func (r *showRepository) Create(show *models.Show) error {
	return r.db.Create(show).Error
}

// GetByID retrieves a show by ID
func (r *showRepository) GetByID(id uint) (*models.Show, error) {
	var show models.Show
	err := r.db.First(&show, id).Error
	if err != nil {
		return nil, err
	}
	return &show, nil
}

// GetByTmdbID retrieves a show by TMDB ID
func (r *showRepository) GetByTmdbID(tmdbID int) (*models.Show, error) {
	var show models.Show
	err := r.db.Where("tmdb_id = ?", tmdbID).First(&show).Error
	if err != nil {
		return nil, err
	}
	return &show, nil
}

// List retrieves shows with pagination
func (r *showRepository) List(page, pageSize int) ([]*models.Show, int64, error) {
	return r.listWithFilters("", "", page, pageSize)
}

// ListByStatus retrieves shows with pagination filtered by status
func (r *showRepository) ListByStatus(status string, page, pageSize int) ([]*models.Show, int64, error) {
	return r.listWithFilters(status, "", page, pageSize)
}

// ListFiltered retrieves shows with pagination filtered by status and search keyword
func (r *showRepository) ListFiltered(status, search string, page, pageSize int) ([]*models.Show, int64, error) {
	return r.listWithFilters(status, search, page, pageSize)
}

// ListAll retrieves all shows
func (r *showRepository) ListAll() ([]*models.Show, error) {
	var shows []*models.Show
	err := r.db.Find(&shows).Error
	return shows, err
}

// ListReturning retrieves all returning/airing shows
func (r *showRepository) ListReturning() ([]*models.Show, error) {
	var shows []*models.Show
	err := r.db.Where("status = ?", "Returning Series").
		Order("next_air_date ASC").
		Find(&shows).Error
	return shows, err
}

// Update updates a show
func (r *showRepository) Update(show *models.Show) error {
	return r.db.Save(show).Error
}

// Delete deletes a show by ID
func (r *showRepository) Delete(id uint) error {
	return r.db.Delete(&models.Show{}, id).Error
}

// Count returns the total number of shows
func (r *showRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.Show{}).Count(&count).Error
	return count, err
}

// Search searches shows by name or original name
func (r *showRepository) Search(query string, page, pageSize int) ([]*models.Show, int64, error) {
	return r.listWithFilters("", query, page, pageSize)
}

func (r *showRepository) listWithFilters(status, search string, page, pageSize int) ([]*models.Show, int64, error) {
	var shows []*models.Show
	var total int64

	query := r.db.Model(&models.Show{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if search != "" {
		if r.db.Dialector.Name() == "sqlite" {
			q := strings.ToLower(search)
			query = query.Where("LOWER(name) LIKE ? OR LOWER(original_name) LIKE ?", "%"+q+"%", "%"+q+"%")
		} else {
			query = query.Where("name ILIKE ? OR original_name ILIKE ?", "%"+search+"%", "%"+search+"%")
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&shows).Error

	return shows, total, err
}
