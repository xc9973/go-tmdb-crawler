package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/xc9973/go-tmdb-crawler/models"
	"github.com/xc9973/go-tmdb-crawler/repositories"
)

// CrawlerService handles crawling operations
type CrawlerService struct {
	tmdb        *TMDBService
	showRepo    repositories.ShowRepository
	episodeRepo repositories.EpisodeRepository
	logRepo     repositories.CrawlLogRepository
	taskRepo    repositories.CrawlTaskRepository
}

// NewCrawlerService creates a new crawler service instance
func NewCrawlerService(
	tmdb *TMDBService,
	showRepo repositories.ShowRepository,
	episodeRepo repositories.EpisodeRepository,
	logRepo repositories.CrawlLogRepository,
	taskRepo repositories.CrawlTaskRepository,
) *CrawlerService {
	return &CrawlerService{
		tmdb:        tmdb,
		showRepo:    showRepo,
		episodeRepo: episodeRepo,
		logRepo:     logRepo,
		taskRepo:    taskRepo,
	}
}

// CrawlResult represents the result of a crawl operation
type CrawlResult struct {
	TmdbID        int
	Success       bool
	EpisodesCount int
	Error         error
	Duration      time.Duration
}

// CrawlShow crawls a single show from TMDB
func (s *CrawlerService) CrawlShow(tmdbID int) error {
	startTime := time.Now()

	// Fetch show details from TMDB
	tmdbShow, err := s.tmdb.GetShowDetails(tmdbID)
	if err != nil {
		s.createCrawlLog(nil, tmdbID, "fetch", "failed", 0, err.Error(), startTime)
		return fmt.Errorf("failed to fetch show details: %w", err)
	}

	// Parse dates
	firstAirDate, _ := ParseDate(tmdbShow.FirstAirDate)

	// Check if show already exists
	show, err := s.showRepo.GetByTmdbID(tmdbID)
	if err != nil {
		// Create new show
		show = &models.Show{
			TmdbID:       tmdbShow.ID,
			Name:         tmdbShow.Name,
			OriginalName: tmdbShow.OriginalName,
			Status:       tmdbShow.Status,
			FirstAirDate: firstAirDate,
			Overview:     tmdbShow.Overview,
			PosterPath:   tmdbShow.PosterPath,
			BackdropPath: tmdbShow.BackdropPath,
			Popularity:   tmdbShow.Popularity,
			VoteAverage:  tmdbShow.VoteAverage,
			VoteCount:    tmdbShow.VoteCount,
		}

		// Parse genres
		if len(tmdbShow.Genres) > 0 {
			genresJSON, _ := json.Marshal(tmdbShow.Genres)
			show.Genres = string(genresJSON)
		}

		if err := s.showRepo.Create(show); err != nil {
			s.createCrawlLog(nil, tmdbID, "fetch", "failed", 0, err.Error(), startTime)
			return fmt.Errorf("failed to create show: %w", err)
		}
	} else {
		// Update existing show
		show.Name = tmdbShow.Name
		show.OriginalName = tmdbShow.OriginalName
		show.Status = tmdbShow.Status
		show.Overview = tmdbShow.Overview
		show.PosterPath = tmdbShow.PosterPath
		show.BackdropPath = tmdbShow.BackdropPath
		show.Popularity = tmdbShow.Popularity
		show.VoteAverage = tmdbShow.VoteAverage
		show.VoteCount = tmdbShow.VoteCount

		if len(tmdbShow.Genres) > 0 {
			genresJSON, _ := json.Marshal(tmdbShow.Genres)
			show.Genres = string(genresJSON)
		}

		if err := s.showRepo.Update(show); err != nil {
			s.createCrawlLog(&show.ID, tmdbID, "fetch", "failed", 0, err.Error(), startTime)
			return fmt.Errorf("failed to update show: %w", err)
		}
	}

	// Crawl all seasons
	totalEpisodes := 0
	for _, season := range tmdbShow.Seasons {
		if season.SeasonNumber == 0 {
			continue // Skip specials
		}

		episodes, err := s.crawlSeason(int(show.ID), tmdbID, season.SeasonNumber)
		if err != nil {
			s.createCrawlLog(&show.ID, tmdbID, "fetch", "partial", totalEpisodes, err.Error(), startTime)
			return fmt.Errorf("failed to crawl season %d: %w", season.SeasonNumber, err)
		}
		totalEpisodes += len(episodes)
	}

	// Update show metadata
	if len(tmdbShow.Seasons) > 0 {
		lastSeason := tmdbShow.Seasons[len(tmdbShow.Seasons)-1]
		show.LastSeasonNumber = lastSeason.SeasonNumber
		show.LastEpisodeCount = lastSeason.EpisodeCount
	}
	show.LastCrawledAt = &[]time.Time{time.Now()}[0]
	s.showRepo.Update(show)

	// Create success log
	s.createCrawlLog(&show.ID, tmdbID, "fetch", "success", totalEpisodes, "", startTime)

	return nil
}

// crawlSeason crawls a specific season
func (s *CrawlerService) crawlSeason(showID, tmdbID, seasonNumber int) ([]*models.Episode, error) {
	// Fetch season details from TMDB
	tmdbSeason, err := s.tmdb.GetSeasonEpisodes(tmdbID, seasonNumber)
	if err != nil {
		return nil, err
	}

	var episodes []*models.Episode

	for _, tmdbEpisode := range tmdbSeason.Episodes {
		airDate, _ := ParseDate(tmdbEpisode.AirDate)

		episode := &models.Episode{
			ShowID:        uint(showID),
			SeasonNumber:  tmdbEpisode.SeasonNumber,
			EpisodeNumber: tmdbEpisode.EpisodeNumber,
			Name:          tmdbEpisode.Name,
			Overview:      tmdbEpisode.Overview,
			AirDate:       airDate,
			StillPath:     tmdbEpisode.StillPath,
			Runtime:       tmdbEpisode.Runtime,
			VoteAverage:   tmdbEpisode.VoteAverage,
			VoteCount:     tmdbEpisode.VoteCount,
		}

		// Check if episode already exists
		// For simplicity, we'll just update/create
		// In production, you might want to check and update instead
		episodes = append(episodes, episode)
	}

	// Batch create/update episodes
	if err := s.episodeRepo.CreateBatch(episodes); err != nil {
		return nil, err
	}

	return episodes, nil
}

// BatchCrawl crawls multiple shows
func (s *CrawlerService) BatchCrawl(tmdbIDs []int) []*CrawlResult {
	results := make([]*CrawlResult, 0, len(tmdbIDs))

	for _, tmdbID := range tmdbIDs {
		startTime := time.Now()
		err := s.CrawlShow(tmdbID)
		duration := time.Since(startTime)

		result := &CrawlResult{
			TmdbID:   tmdbID,
			Success:  err == nil,
			Error:    err,
			Duration: duration,
		}

		if err == nil {
			// Get episode count
			if show, err := s.showRepo.GetByTmdbID(tmdbID); err == nil {
				if count, err := s.episodeRepo.CountByShowID(show.ID); err == nil {
					result.EpisodesCount = int(count)
				}
			}
		}

		results = append(results, result)
	}

	return results
}

// RefreshAll refreshes all shows in the database
func (s *CrawlerService) RefreshAll() error {
	shows, err := s.showRepo.ListAll()
	if err != nil {
		return fmt.Errorf("failed to list shows: %w", err)
	}

	tmdbIDs := make([]int, len(shows))
	for i, show := range shows {
		tmdbIDs[i] = show.TmdbID
	}

	results := s.BatchCrawl(tmdbIDs)

	// Check if all succeeded
	for _, result := range results {
		if !result.Success {
			return fmt.Errorf("failed to crawl show %d: %w", result.TmdbID, result.Error)
		}
	}

	return nil
}

// CrawlByStatus refreshes shows based on status filter
func (s *CrawlerService) CrawlByStatus(status string) error {
	var shows []*models.Show
	var err error

	if status == "returning" || status == "Returning Series" {
		shows, err = s.showRepo.ListReturning()
	} else {
		shows, err = s.showRepo.ListAll()
	}
	if err != nil {
		return fmt.Errorf("failed to list shows: %w", err)
	}

	tmdbIDs := make([]int, len(shows))
	for i, show := range shows {
		tmdbIDs[i] = show.TmdbID
	}

	results := s.BatchCrawl(tmdbIDs)
	for _, result := range results {
		if !result.Success {
			return fmt.Errorf("failed to crawl show %d: %w", result.TmdbID, result.Error)
		}
	}

	return nil
}

// createCrawlLog creates a crawl log entry
func (s *CrawlerService) createCrawlLog(showID *uint, tmdbID int, action, status string, episodesCount int, errorMsg string, startTime time.Time) {
	duration := time.Since(startTime)

	log := &models.CrawlLog{
		ShowID:        showID,
		TmdbID:        tmdbID,
		Action:        action,
		Status:        status,
		EpisodesCount: episodesCount,
		ErrorMessage:  errorMsg,
		DurationMs:    int(duration.Milliseconds()),
	}

	_ = s.logRepo.Create(log)
}
