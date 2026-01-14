package backup

import (
	"fmt"

	"github.com/xc9973/go-tmdb-crawler/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Import imports data from a BackupExport structure
func (s *service) Import(backup *models.BackupExport, mode ImportMode) (*models.ImportResult, error) {
	// Validate version
	if backup.Version != models.BackupVersion {
		return nil, fmt.Errorf("unsupported backup version: %s (supported: %s)",
			backup.Version, models.BackupVersion)
	}

	// Start transaction
	tx := s.db.Begin()
	txCommitted := false
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		} else if !txCommitted {
			tx.Rollback()
		}
	}()

	result := &models.ImportResult{}

	if mode == ImportModeReplace {
		// Clear all tables in correct order (respecting foreign keys)
		if err := s.clearAllTables(tx); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// Import in order to respect foreign key dependencies
	// 1. Shows (master table)
	showsImported, conflictsSkipped, err := s.importShows(tx, backup.Data.Shows, mode)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to import shows: %w", err)
	}
	result.ShowsImported = showsImported
	result.ConflictsSkipped += conflictsSkipped

	// 2. TelegraphPosts (no foreign key to shows)
	telegraphPostsImported, conflictsSkipped, err := s.importTelegraphPosts(tx, backup.Data.TelegraphPosts, mode)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to import telegraph posts: %w", err)
	}
	result.TelegraphPostsImported = telegraphPostsImported
	result.ConflictsSkipped += conflictsSkipped

	// 3. Episodes (depends on show_id)
	episodesImported, conflictsSkipped, err := s.importEpisodes(tx, backup.Data.Episodes, mode)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to import episodes: %w", err)
	}
	result.EpisodesImported = episodesImported
	result.ConflictsSkipped += conflictsSkipped

	// 4. CrawlLogs (depends on show_id)
	crawlLogsImported, conflictsSkipped, err := s.importCrawlLogs(tx, backup.Data.CrawlLogs, mode)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to import crawl logs: %w", err)
	}
	result.CrawlLogsImported = crawlLogsImported
	result.ConflictsSkipped += conflictsSkipped

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	txCommitted = true

	return result, nil
}

// clearAllTables clears all tables in correct order
func (s *service) clearAllTables(tx *gorm.DB) error {
	// Delete in reverse dependency order
	if err := tx.Where("1 = 1").Delete(&models.CrawlLog{}).Error; err != nil {
		return fmt.Errorf("failed to clear crawl_logs: %w", err)
	}
	if err := tx.Where("1 = 1").Delete(&models.Episode{}).Error; err != nil {
		return fmt.Errorf("failed to clear episodes: %w", err)
	}
	if err := tx.Where("1 = 1").Delete(&models.TelegraphPost{}).Error; err != nil {
		return fmt.Errorf("failed to clear telegraph_posts: %w", err)
	}
	if err := tx.Where("1 = 1").Delete(&models.Show{}).Error; err != nil {
		return fmt.Errorf("failed to clear shows: %w", err)
	}
	return nil
}

// importShows imports shows with explicit ID handling
func (s *service) importShows(tx *gorm.DB, shows []models.Show, mode ImportMode) (int, int, error) {
	if len(shows) == 0 {
		return 0, 0, nil
	}

	if mode == ImportModeReplace {
		// Direct import with IDs (table is empty)
		for _, show := range shows {
			if err := tx.Create(&show).Error; err != nil {
				return 0, 0, err
			}
		}
		return len(shows), 0, nil
	}

	// Merge mode: use OnConflict to handle duplicates
	imported := 0
	conflicts := 0

	for _, show := range shows {
		result := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoNothing: true,
		}).Create(&show)

		if result.Error != nil {
			return 0, 0, result.Error
		}
		if result.RowsAffected == 0 {
			conflicts++
		} else {
			imported++
		}
	}

	return imported, conflicts, nil
}

// importEpisodes imports episodes with conflict handling
func (s *service) importEpisodes(tx *gorm.DB, episodes []models.Episode, mode ImportMode) (int, int, error) {
	if len(episodes) == 0 {
		return 0, 0, nil
	}

	if mode == ImportModeReplace {
		// Direct import with IDs (table is empty)
		for _, ep := range episodes {
			if err := tx.Create(&ep).Error; err != nil {
				return 0, 0, err
			}
		}
		return len(episodes), 0, nil
	}

	// Merge mode: use OnConflict to handle duplicates
	imported := 0
	conflicts := 0

	for _, ep := range episodes {
		result := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "show_id"}, {Name: "season_number"}, {Name: "episode_number"}},
			DoNothing: true,
		}).Create(&ep)

		if result.Error != nil {
			return 0, 0, result.Error
		}
		if result.RowsAffected == 0 {
			conflicts++
		} else {
			imported++
		}
	}

	return imported, conflicts, nil
}

// importCrawlLogs imports crawl logs
func (s *service) importCrawlLogs(tx *gorm.DB, logs []models.CrawlLog, mode ImportMode) (int, int, error) {
	if len(logs) == 0 {
		return 0, 0, nil
	}

	if mode == ImportModeReplace {
		for _, log := range logs {
			if err := tx.Create(&log).Error; err != nil {
				return 0, 0, err
			}
		}
		return len(logs), 0, nil
	}

	imported := 0
	conflicts := 0

	for _, log := range logs {
		result := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoNothing: true,
		}).Create(&log)

		if result.Error != nil {
			return 0, 0, result.Error
		}
		if result.RowsAffected == 0 {
			conflicts++
		} else {
			imported++
		}
	}

	return imported, conflicts, nil
}

// importTelegraphPosts imports telegraph posts
func (s *service) importTelegraphPosts(tx *gorm.DB, posts []models.TelegraphPost, mode ImportMode) (int, int, error) {
	if len(posts) == 0 {
		return 0, 0, nil
	}

	if mode == ImportModeReplace {
		for _, post := range posts {
			if err := tx.Create(&post).Error; err != nil {
				return 0, 0, err
			}
		}
		return len(posts), 0, nil
	}

	imported := 0
	conflicts := 0

	for _, post := range posts {
		result := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoNothing: true,
		}).Create(&post)

		if result.Error != nil {
			return 0, 0, result.Error
		}
		if result.RowsAffected == 0 {
			conflicts++
		} else {
			imported++
		}
	}

	return imported, conflicts, nil
}

// GetStatus returns the current backup status
func (s *service) GetStatus() (*models.BackupStatus, error) {
	showsCount, err := s.showRepo.Count()
	if err != nil {
		return nil, fmt.Errorf("failed to count shows: %w", err)
	}

	episodesCount, err := s.episodeRepo.Count()
	if err != nil {
		return nil, fmt.Errorf("failed to count episodes: %w", err)
	}

	crawlLogsCount, err := s.crawlLogRepo.Count()
	if err != nil {
		return nil, fmt.Errorf("failed to count crawl logs: %w", err)
	}

	telegraphPostsCount, err := s.telegraphPostRepo.Count()
	if err != nil {
		return nil, fmt.Errorf("failed to count telegraph posts: %w", err)
	}

	return &models.BackupStatus{
		Stats: models.BackupStats{
			Shows:          int(showsCount),
			Episodes:       int(episodesCount),
			CrawlLogs:      int(crawlLogsCount),
			TelegraphPosts: int(telegraphPostsCount),
		},
	}, nil
}
