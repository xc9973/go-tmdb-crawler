package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/xc9973/go-tmdb-crawler/utils"
)

// Scheduler handles scheduled tasks
type Scheduler struct {
	cron            *cron.Cron
	crawler         *CrawlerService
	publisher       *PublisherService
	logger          *utils.Logger
	mu              sync.RWMutex
	running         bool
	lastCrawlTime   time.Time
	lastPublishTime time.Time
}

// NewScheduler creates a new scheduler instance
func NewScheduler(
	crawler *CrawlerService,
	publisher *PublisherService,
	logger *utils.Logger,
) *Scheduler {
	return &Scheduler{
		cron:      cron.New(cron.WithSeconds()),
		crawler:   crawler,
		publisher: publisher,
		logger:    logger,
		running:   false,
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("scheduler is already running")
	}

	s.logger.Info("Starting scheduler...")

	// Add jobs
	if _, err := s.cron.AddFunc("0 0 8,12,20 * * *", s.dailyCrawlJob); err != nil {
		return fmt.Errorf("failed to add daily crawl job: %w", err)
	}

	if _, err := s.cron.AddFunc("0 30 20 * * *", s.dailyPublishJob); err != nil {
		return fmt.Errorf("failed to add daily publish job: %w", err)
	}

	if _, err := s.cron.AddFunc("0 0 6 * * 1", s.weeklyCrawlJob); err != nil {
		return fmt.Errorf("failed to add weekly crawl job: %w", err)
	}

	if _, err := s.cron.AddFunc("0 0 7 * * 1", s.weeklyPublishJob); err != nil {
		return fmt.Errorf("failed to add weekly publish job: %w", err)
	}

	s.cron.Start()
	s.running = true

	s.logger.Info("Scheduler started successfully")
	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.logger.Info("Stopping scheduler...")
	s.cron.Stop()
	s.running = false
	s.logger.Info("Scheduler stopped")
}

// IsRunning checks if the scheduler is running
func (s *Scheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// dailyCrawlJob performs daily crawl task
func (s *Scheduler) dailyCrawlJob() {
	s.logger.Info("Starting daily crawl job...")
	startTime := time.Now()

	// Refresh all returning shows
	go func() {
		if err := s.crawler.RefreshAll(); err != nil {
			s.logger.Error("Daily crawl failed: %v", err)
		} else {
			s.mu.Lock()
			s.lastCrawlTime = time.Now()
			s.mu.Unlock()
			duration := time.Since(startTime)
			s.logger.Info("Daily crawl completed in %v", duration)
		}
	}()
}

// dailyPublishJob performs daily publish task
func (s *Scheduler) dailyPublishJob() {
	s.logger.Info("Starting daily publish job...")
	startTime := time.Now()

	// Publish today's updates
	go func() {
		result, err := s.publisher.PublishTodayUpdates()
		if err != nil {
			s.logger.Error("Daily publish failed: %v", err)
		} else if result.Success {
			s.mu.Lock()
			s.lastPublishTime = time.Now()
			s.mu.Unlock()
			duration := time.Since(startTime)
			s.logger.Info("Daily publish completed: %s (%d shows, %d episodes) in %v",
				result.URL,
				result.ShowsCount,
				result.EpisodesCount,
				duration)
		} else {
			s.logger.Warn("Daily publish skipped: %v", result.Error)
		}
	}()
}

// weeklyCrawlJob performs weekly full crawl
func (s *Scheduler) weeklyCrawlJob() {
	s.logger.Info("Starting weekly crawl job...")
	startTime := time.Now()

	// Refresh all shows
	go func() {
		if err := s.crawler.RefreshAll(); err != nil {
			s.logger.Error("Weekly crawl failed: %v", err)
		} else {
			s.mu.Lock()
			s.lastCrawlTime = time.Now()
			s.mu.Unlock()
			duration := time.Since(startTime)
			s.logger.Info("Weekly crawl completed in %v", duration)
		}
	}()
}

// weeklyPublishJob performs weekly publish
func (s *Scheduler) weeklyPublishJob() {
	s.logger.Info("Starting weekly publish job...")
	startTime := time.Now()

	// Publish weekly updates
	go func() {
		result, err := s.publisher.PublishWeeklyUpdates()
		if err != nil {
			s.logger.Error("Weekly publish failed: %v", err)
		} else if result.Success {
			s.mu.Lock()
			s.lastPublishTime = time.Now()
			s.mu.Unlock()
			duration := time.Since(startTime)
			s.logger.Info("Weekly publish completed: %s (%d shows, %d episodes) in %v",
				result.URL,
				result.ShowsCount,
				result.EpisodesCount,
				duration)
		} else {
			s.logger.Warn("Weekly publish skipped: %v", result.Error)
		}
	}()
}

// RunCrawlNow triggers an immediate crawl job
func (s *Scheduler) RunCrawlNow() error {
	s.logger.Info("Triggering immediate crawl...")
	startTime := time.Now()

	if err := s.crawler.RefreshAll(); err != nil {
		return fmt.Errorf("crawl failed: %w", err)
	}

	s.mu.Lock()
	s.lastCrawlTime = time.Now()
	s.mu.Unlock()

	duration := time.Since(startTime)
	s.logger.Info("Immediate crawl completed in %v", duration)
	return nil
}

// RunPublishNow triggers an immediate publish job
func (s *Scheduler) RunPublishNow() (*PublishResult, error) {
	s.logger.Info("Triggering immediate publish...")
	startTime := time.Now()

	result, err := s.publisher.PublishTodayUpdates()
	if err != nil {
		return nil, fmt.Errorf("publish failed: %w", err)
	}

	if result.Success {
		s.mu.Lock()
		s.lastPublishTime = time.Now()
		s.mu.Unlock()

		duration := time.Since(startTime)
		s.logger.Info("Immediate publish completed: %s (%d shows, %d episodes) in %v",
			result.URL,
			result.ShowsCount,
			result.EpisodesCount,
			duration)
	}

	return result, nil
}

// GetStatus returns the scheduler status
func (s *Scheduler) GetStatus() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := map[string]interface{}{
		"running":           s.running,
		"last_crawl_time":   s.lastCrawlTime,
		"last_publish_time": s.lastPublishTime,
	}

	if !s.lastCrawlTime.IsZero() {
		status["time_since_last_crawl"] = time.Since(s.lastCrawlTime).String()
	}

	if !s.lastPublishTime.IsZero() {
		status["time_since_last_publish"] = time.Since(s.lastPublishTime).String()
	}

	return status
}

// GetNextRunTimes returns the next scheduled run times
func (s *Scheduler) GetNextRunTimes() map[string]string {
	entries := s.cron.Entries()
	result := make(map[string]string)

	for _, entry := range entries {
		next := entry.Next.Format("2006-01-02 15:04:05")
		result[fmt.Sprintf("job_%d", entry.ID)] = next
	}

	return result
}

// AddCustomJob adds a custom cron job
func (s *Scheduler) AddCustomJob(spec string, job func()) (cron.EntryID, error) {
	return s.cron.AddFunc(spec, job)
}

// RemoveJob removes a job by ID
func (s *Scheduler) RemoveJob(id cron.EntryID) {
	s.cron.Remove(id)
}

// RunManualCrawl runs a manual crawl task
func (s *Scheduler) RunManualCrawl(showID int) error {
	s.logger.Info("Running manual crawl for show %d", showID)

	if err := s.crawler.CrawlShow(showID); err != nil {
		return fmt.Errorf("manual crawl failed: %w", err)
	}

	s.logger.Info("Manual crawl completed for show %d", showID)
	return nil
}

// RunManualPublish runs a manual publish task
func (s *Scheduler) RunManualPublish(showID uint) (*PublishResult, error) {
	s.logger.Info("Running manual publish for show %d", showID)

	result, err := s.publisher.PublishShow(showID)
	if err != nil {
		return nil, fmt.Errorf("manual publish failed: %w", err)
	}

	if result.Success {
		s.mu.Lock()
		s.lastPublishTime = time.Now()
		s.mu.Unlock()

		s.logger.Info("Manual publish completed: %s", result.URL)
	}

	return result, nil
}

// SetCronSpec updates cron specification for a job type
func (s *Scheduler) SetCronSpec(jobType, spec string) error {
	// This would require stopping the scheduler and restarting
	// For now, just log a warning
	s.logger.Warn("SetCronSpec not implemented - restart scheduler to change specs")
	return fmt.Errorf("not implemented")
}

// ValidateCronSpec validates a cron specification
func ValidateCronSpec(spec string) error {
	_, err := cron.ParseStandard(spec)
	return err
}

// GetDefaultCronSpecs returns default cron specifications
func GetDefaultCronSpecs() map[string]string {
	return map[string]string{
		"daily_crawl":    "0 0 8,12,20 * * *", // 8am, 12pm, 8pm
		"daily_publish":  "0 30 20 * * *",     // 8:30pm
		"weekly_crawl":   "0 0 6 * * 1",       // Monday 6am
		"weekly_publish": "0 0 7 * * 1",       // Monday 7am
	}
}

// Helper function to run job with error handling
func (s *Scheduler) runJobWithErrorHandling(jobName string, job func() error) {
	s.logger.Info("Starting %s", jobName)
	startTime := time.Now()

	if err := job(); err != nil {
		s.logger.Error("%s failed: %v", jobName, err)
	} else {
		duration := time.Since(startTime)
		s.logger.Info("%s completed in %v", jobName, duration)
	}
}
