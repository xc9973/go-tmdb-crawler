package backup

import (
	"github.com/xc9973/go-tmdb-crawler/models"
	"github.com/xc9973/go-tmdb-crawler/repositories"
	"gorm.io/gorm"
)

// ImportMode defines the import mode
type ImportMode string

const (
	ImportModeReplace ImportMode = "replace" // Clear all tables then import
	ImportModeMerge   ImportMode = "merge"   // Skip existing records by ID
)

// Service handles backup export and import operations
type Service interface {
	// Export exports all data to a BackupExport structure
	Export() (*models.BackupExport, error)

	// Import imports data from a BackupExport structure
	Import(backup *models.BackupExport, mode ImportMode) (*models.ImportResult, error)

	// GetStatus returns the current backup status
	GetStatus() (*models.BackupStatus, error)
}

type service struct {
	db                *gorm.DB
	showRepo          repositories.ShowRepository
	episodeRepo       repositories.EpisodeRepository
	crawlLogRepo      repositories.CrawlLogRepository
	telegraphPostRepo repositories.TelegraphPostRepository
}

// NewService creates a new backup service
func NewService(
	db *gorm.DB,
	showRepo repositories.ShowRepository,
	episodeRepo repositories.EpisodeRepository,
	crawlLogRepo repositories.CrawlLogRepository,
	telegraphPostRepo repositories.TelegraphPostRepository,
) Service {
	return &service{
		db:                db,
		showRepo:          showRepo,
		episodeRepo:       episodeRepo,
		crawlLogRepo:      crawlLogRepo,
		telegraphPostRepo: telegraphPostRepo,
	}
}
