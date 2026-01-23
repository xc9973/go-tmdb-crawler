package correction

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/xc9973/go-tmdb-crawler/models"
	"github.com/xc9973/go-tmdb-crawler/repositories"
)

// Crawler defines the interface for crawler operations
type Crawler interface {
	CrawlShow(tmdbID int) error
}

// Service orchestrates the correction detection and refresh process
type Service struct {
	showRepo    repositories.ShowRepository
	episodeRepo repositories.EpisodeRepository
	taskRepo    repositories.CrawlTaskRepository
	crawler     Crawler
	detector    *Detector
	lastResult  *DetectionResult
	resultMutex sync.RWMutex
}

// NewService creates a new correction service
func NewService(
	showRepo repositories.ShowRepository,
	episodeRepo repositories.EpisodeRepository,
	taskRepo repositories.CrawlTaskRepository,
	crawler Crawler,
	location *time.Location,
) *Service {
	return &Service{
		showRepo:    showRepo,
		episodeRepo: episodeRepo,
		taskRepo:    taskRepo,
		crawler:     crawler,
		detector:    NewDetector(location),
	}
}

// GetLastDetectionResult returns the cached detection result
func (s *Service) GetLastDetectionResult() *DetectionResult {
	s.resultMutex.RLock()
	defer s.resultMutex.RUnlock()
	return s.lastResult
}

// DetectionResult contains statistics from a detection run
type DetectionResult struct {
	TotalShowsAnalyzed int
	StaleShowsFound    int
	TasksCreated       int
	Duration           time.Duration
	StaleShows         []*StaleShowInfo
}

// RunDetection analyzes all shows and creates correction tasks for stale ones
func (s *Service) RunDetection() (*DetectionResult, error) {
	startTime := time.Now()

	// Get all shows (could optimize to only get returning/ended)
	shows, err := s.showRepo.ListAll()
	if err != nil {
		return nil, fmt.Errorf("failed to list shows: %w", err)
	}

	result := &DetectionResult{
		TotalShowsAnalyzed: len(shows),
		StaleShows:         make([]*StaleShowInfo, 0),
	}

	// Analyze each show
	for _, show := range shows {
		staleInfo, err := s.analyzeShow(show)
		if err != nil {
			continue // Log error but continue with other shows
		}

		if staleInfo != nil {
			result.StaleShows = append(result.StaleShows, staleInfo)
		}
	}

	result.StaleShowsFound = len(result.StaleShows)

	// Create correction tasks for stale shows
	for _, stale := range result.StaleShows {
		if err := s.createCorrectionTask(stale); err != nil {
			// Log error but continue
			continue
		}
		result.TasksCreated++
	}

	result.Duration = time.Since(startTime)

	// Cache the result
	s.resultMutex.Lock()
	s.lastResult = result
	s.resultMutex.Unlock()

	return result, nil
}

// analyzeShow checks if a single show is stale
func (s *Service) analyzeShow(show *models.Show) (*StaleShowInfo, error) {
	// Get episodes for this show
	episodes, err := s.episodeRepo.GetByShowID(show.ID)
	if err != nil {
		return nil, err
	}

	// Need at least 3 episodes
	if len(episodes) < 3 {
		return nil, nil
	}

	// Extract air dates
	dates := make([]time.Time, 0, len(episodes))
	for _, ep := range episodes {
		if ep.AirDate != nil {
			dates = append(dates, *ep.AirDate)
		}
	}

	// Sort dates chronologically
	if len(dates) > 0 {
		sort.Slice(dates, func(i, j int) bool {
			return dates[i].Before(dates[j])
		})
	}

	// Detect staleness
	var customThreshold *int
	if show.RefreshThreshold > 0 {
		customThreshold = &show.RefreshThreshold
	}

	staleInfo := s.detector.DetectStale(
		show.ID,
		show.TmdbID,
		show.Name,
		dates,
		customThreshold,
	)

	return staleInfo, nil
}

// createCorrectionTask creates a crawl task for refreshing a stale show
func (s *Service) createCorrectionTask(stale *StaleShowInfo) error {
	now := time.Now()

	// Update show's stale detection timestamp
	show, err := s.showRepo.GetByID(stale.ShowID)
	if err != nil {
		return fmt.Errorf("failed to get show: %w", err)
	}

	show.StaleDetectedAt = &now
	show.LastCorrectionResult = fmt.Sprintf("Detected: %d days overdue", stale.DaysOverdue)
	if err := s.showRepo.Update(show); err != nil {
		return fmt.Errorf("failed to update show: %w", err)
	}

	// Create correction task
	task := &models.CrawlTask{
		Type:      "correction",
		Status:    "queued",
		Params:    fmt.Sprintf(`{"show_id": %d, "tmdb_id": %d}`, stale.ShowID, stale.TmdbID),
		CreatedAt: now,
	}

	if err := s.taskRepo.Create(task); err != nil {
		return fmt.Errorf("failed to create correction task: %w", err)
	}

	return nil
}

// RefreshShow manually refreshes a specific show (for immediate correction)
func (s *Service) RefreshShow(showID uint, tmdbID int) error {
	return s.crawler.CrawlShow(tmdbID)
}

// ClearStaleFlag removes the stale_detected_at flag from a show
func (s *Service) ClearStaleFlag(showID uint) error {
	show, err := s.showRepo.GetByID(showID)
	if err != nil {
		return err
	}

	show.StaleDetectedAt = nil
	show.LastCorrectionResult = ""
	return s.showRepo.Update(show)
}

// SetCustomThreshold sets a custom refresh threshold for a show
func (s *Service) SetCustomThreshold(showID uint, threshold int) error {
	show, err := s.showRepo.GetByID(showID)
	if err != nil {
		return err
	}

	show.RefreshThreshold = threshold
	return s.showRepo.Update(show)
}
