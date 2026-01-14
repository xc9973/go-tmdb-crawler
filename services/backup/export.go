package backup

import (
	"time"

	"github.com/xc9973/go-tmdb-crawler/models"
)

const appVersion = "2.0.0"

// Export exports all data to a BackupExport structure
func (s *service) Export() (*models.BackupExport, error) {
	// Fetch all shows
	shows, err := s.showRepo.ListAll()
	if err != nil {
		return nil, err
	}

	// Build episodes list by show
	var allEpisodes []models.Episode
	for _, show := range shows {
		episodes, err := s.episodeRepo.GetByShowID(show.ID)
		if err != nil {
			return nil, err
		}
		for _, ep := range episodes {
			allEpisodes = append(allEpisodes, *ep)
		}
	}

	// Fetch crawl logs
	crawlLogs, err := s.crawlLogRepo.ListAll()
	if err != nil {
		return nil, err
	}

	// Fetch telegraph posts
	telegraphPosts, err := s.telegraphPostRepo.ListAll()
	if err != nil {
		return nil, err
	}

	// Build stats
	stats := models.BackupStats{
		Shows:          len(shows),
		Episodes:       len(allEpisodes),
		CrawlLogs:      len(crawlLogs),
		TelegraphPosts: len(telegraphPosts),
	}

	// Build backup data
	data := models.BackupData{
		Shows:          convertShowsToSlice(shows),
		Episodes:       allEpisodes,
		CrawlLogs:      convertCrawlLogsToSlice(crawlLogs),
		TelegraphPosts: convertTelegraphPostsToSlice(telegraphPosts),
	}

	return &models.BackupExport{
		Version:    models.BackupVersion,
		ExportedAt: time.Now(),
		AppVersion: appVersion,
		Stats:      stats,
		Data:       data,
	}, nil
}

// Helper functions to convert pointer slices to value slices
func convertShowsToSlice(shows []*models.Show) []models.Show {
	result := make([]models.Show, len(shows))
	for i, s := range shows {
		result[i] = *s
	}
	return result
}

func convertCrawlLogsToSlice(logs []*models.CrawlLog) []models.CrawlLog {
	result := make([]models.CrawlLog, len(logs))
	for i, l := range logs {
		result[i] = *l
	}
	return result
}

func convertTelegraphPostsToSlice(posts []*models.TelegraphPost) []models.TelegraphPost {
	result := make([]models.TelegraphPost, len(posts))
	for i, p := range posts {
		result[i] = *p
	}
	return result
}
