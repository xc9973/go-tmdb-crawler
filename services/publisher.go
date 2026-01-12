package services

import (
	"fmt"
	"time"

	"github.com/xc9973/go-tmdb-crawler/repositories"
	"github.com/xc9973/go-tmdb-crawler/utils"
)

// PublisherService handles publishing to Telegraph
type PublisherService struct {
	telegraph      *TelegraphService
	showRepo       repositories.ShowRepository
	episodeRepo    repositories.EpisodeRepository
	timezoneHelper *utils.TimezoneHelper
}

// NewPublisherService creates a new publisher service instance
func NewPublisherService(
	telegraph *TelegraphService,
	showRepo repositories.ShowRepository,
	episodeRepo repositories.EpisodeRepository,
	timezoneHelper *utils.TimezoneHelper,
) *PublisherService {
	return &PublisherService{
		telegraph:      telegraph,
		showRepo:       showRepo,
		episodeRepo:    episodeRepo,
		timezoneHelper: timezoneHelper,
	}
}

// PublishResult represents the result of a publish operation
type PublishResult struct {
	Success       bool
	URL           string
	Path          string
	Title         string
	ShowsCount    int
	EpisodesCount int
	Error         error
}

// PublishTodayUpdates publishes today's episode updates to Telegraph
func (s *PublisherService) PublishTodayUpdates() (*PublishResult, error) {
	// Get today's episodes
	episodes, err := s.episodeRepo.GetTodayUpdates()
	if err != nil {
		return &PublishResult{
			Success: false,
			Error:   fmt.Errorf("failed to get today's episodes: %w", err),
		}, err
	}

	if len(episodes) == 0 {
		return &PublishResult{
			Success: false,
			Error:   fmt.Errorf("no episodes found for today"),
		}, fmt.Errorf("no episodes to publish")
	}

	// Generate title using configured timezone
	today := s.timezoneHelper.NowInLocation().Format("2006-01-02")
	title := fmt.Sprintf("今日更新 - %s", today)

	// Generate content
	content := s.telegraph.GenerateUpdateListContent(episodes)

	// Generate tags
	tags := []string{"剧集", "更新", "TV Shows", today}

	// Create page
	page, err := s.telegraph.CreatePage(title, content, tags)
	if err != nil {
		return &PublishResult{
			Success: false,
			Error:   fmt.Errorf("failed to create page: %w", err),
		}, err
	}

	// Count unique shows
	showMap := make(map[uint]bool)
	for _, ep := range episodes {
		showMap[ep.ShowID] = true
	}

	return &PublishResult{
		Success:       true,
		URL:           page.URL,
		Path:          page.Path,
		Title:         title,
		ShowsCount:    len(showMap),
		EpisodesCount: len(episodes),
	}, nil
}

// PublishDateRange publishes episodes for a date range
func (s *PublisherService) PublishDateRange(startDate, endDate time.Time) (*PublishResult, error) {
	// Get episodes in date range
	episodes, err := s.episodeRepo.GetByDateRange(startDate, endDate)
	if err != nil {
		return &PublishResult{
			Success: false,
			Error:   fmt.Errorf("failed to get episodes: %w", err),
		}, err
	}

	if len(episodes) == 0 {
		return &PublishResult{
			Success: false,
			Error:   fmt.Errorf("no episodes found in date range"),
		}, fmt.Errorf("no episodes to publish")
	}

	// Generate title
	title := fmt.Sprintf("更新清单 - %s 至 %s",
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"))

	// Generate content
	content := s.telegraph.GenerateUpdateListContent(episodes)

	// Generate tags
	tags := []string{"剧集", "更新", "TV Shows"}

	// Create page
	page, err := s.telegraph.CreatePage(title, content, tags)
	if err != nil {
		return &PublishResult{
			Success: false,
			Error:   fmt.Errorf("failed to create page: %w", err),
		}, err
	}

	// Count unique shows
	showMap := make(map[uint]bool)
	for _, ep := range episodes {
		showMap[ep.ShowID] = true
	}

	return &PublishResult{
		Success:       true,
		URL:           page.URL,
		Path:          page.Path,
		Title:         title,
		ShowsCount:    len(showMap),
		EpisodesCount: len(episodes),
	}, nil
}

// PublishShow publishes a single show with all its episodes
func (s *PublisherService) PublishShow(showID uint) (*PublishResult, error) {
	// Get show
	show, err := s.showRepo.GetByID(showID)
	if err != nil {
		return &PublishResult{
			Success: false,
			Error:   fmt.Errorf("failed to get show: %w", err),
		}, err
	}

	// Get episodes
	episodes, err := s.episodeRepo.GetByShowID(showID)
	if err != nil {
		return &PublishResult{
			Success: false,
			Error:   fmt.Errorf("failed to get episodes: %w", err),
		}, err
	}

	if len(episodes) == 0 {
		return &PublishResult{
			Success: false,
			Error:   fmt.Errorf("no episodes found for show"),
		}, fmt.Errorf("no episodes to publish")
	}

	// Generate title
	title := fmt.Sprintf("%s - 剧集列表", show.Name)

	// Generate content
	content := s.telegraph.GenerateShowContent(show, episodes)

	// Generate tags
	tags := []string{"剧集", show.Name, "TV Shows"}
	if show.Status != "" {
		tags = append(tags, show.Status)
	}

	// Create page
	page, err := s.telegraph.CreatePage(title, content, tags)
	if err != nil {
		return &PublishResult{
			Success: false,
			Error:   fmt.Errorf("failed to create page: %w", err),
		}, err
	}

	return &PublishResult{
		Success:       true,
		URL:           page.URL,
		Path:          page.Path,
		Title:         title,
		ShowsCount:    1,
		EpisodesCount: len(episodes),
	}, nil
}

// PublishWeeklyUpdates publishes the last 7 days of updates
// Uses the configured timezone for date calculations
func (s *PublisherService) PublishWeeklyUpdates() (*PublishResult, error) {
	today := s.timezoneHelper.TodayInLocation()
	startDate := today.AddDate(0, 0, -7)

	return s.PublishDateRange(startDate, today)
}

// PublishMonthlyUpdates publishes the last 30 days of updates
// Uses the configured timezone for date calculations
func (s *PublisherService) PublishMonthlyUpdates() (*PublishResult, error) {
	today := s.timezoneHelper.TodayInLocation()
	startDate := today.AddDate(0, 0, -30)

	return s.PublishDateRange(startDate, today)
}
