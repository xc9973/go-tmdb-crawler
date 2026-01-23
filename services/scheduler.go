package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/xc9973/go-tmdb-crawler/services/correction"
	"github.com/xc9973/go-tmdb-crawler/utils"
)

// Scheduler handles scheduled tasks
type Scheduler struct {
	cron            *cron.Cron
	crawler         *CrawlerService
	publisher       *PublisherService
	correction      *correction.Service
	logger          *utils.Logger
	mu              sync.RWMutex
	running         bool
	lastCrawlTime   time.Time
	lastPublishTime time.Time

	// Concurrency control
	crawlJobRunning     bool
	publishJobRunning   bool
	correctionJobRunning bool
	crawlJobMutex       sync.Mutex
	publishJobMutex     sync.Mutex
	correctionJobMutex  sync.Mutex

	// Timeout settings
	crawlTimeout   time.Duration
	publishTimeout time.Duration
}

// NewScheduler creates a new scheduler instance
func NewScheduler(
	crawler *CrawlerService,
	publisher *PublisherService,
	correction *correction.Service,
	logger *utils.Logger,
) *Scheduler {
	return &Scheduler{
		cron:           cron.New(cron.WithSeconds()),
		crawler:        crawler,
		publisher:      publisher,
		correction:     correction,
		logger:         logger,
		running:        false,
		crawlTimeout:   30 * time.Minute, // Default crawl timeout
		publishTimeout: 10 * time.Minute, // Default publish timeout
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

	if _, err := s.cron.AddFunc("0 0 2 * * *", s.dailyCorrectionJob); err != nil {
		return fmt.Errorf("failed to add daily correction job: %w", err)
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
	// Check if crawl job is already running
	if !s.crawlJobMutex.TryLock() {
		s.logger.Warn("Daily crawl job already running, skipping")
		return
	}
	defer s.crawlJobMutex.Unlock()

	s.logger.Info("Starting daily crawl job...")
	startTime := time.Now()

	// Refresh all returning shows
	if err := s.crawler.RefreshAll(); err != nil {
		s.logger.Errorf("Daily crawl failed: %v", err)
	} else {
		s.mu.Lock()
		s.lastCrawlTime = time.Now()
		s.mu.Unlock()
		duration := time.Since(startTime)
		s.logger.Infof("Daily crawl completed in %v", duration)
	}
}

// dailyPublishJob performs daily publish task
func (s *Scheduler) dailyPublishJob() {
	// Check if publish job is already running
	if !s.publishJobMutex.TryLock() {
		s.logger.Warn("Daily publish job already running, skipping")
		return
	}
	defer s.publishJobMutex.Unlock()

	s.logger.Info("Starting daily publish job...")
	startTime := time.Now()

	// Publish today's updates
	result, err := s.publisher.PublishTodayUpdates()
	if err != nil {
		s.logger.Errorf("Daily publish failed: %v", err)
	} else if result.Success {
		s.mu.Lock()
		s.lastPublishTime = time.Now()
		s.mu.Unlock()
		duration := time.Since(startTime)
		s.logger.Infof("Daily publish completed: %s (%d shows, %d episodes) in %v",
			result.URL,
			result.ShowsCount,
			result.EpisodesCount,
			duration)
	} else {
		s.logger.Warnf("Daily publish skipped: %v", result.Error)
	}
}

// weeklyCrawlJob performs weekly full crawl
func (s *Scheduler) weeklyCrawlJob() {
	// Check if crawl job is already running
	if !s.crawlJobMutex.TryLock() {
		s.logger.Warn("Weekly crawl job already running, skipping")
		return
	}
	defer s.crawlJobMutex.Unlock()

	s.logger.Info("Starting weekly crawl job...")
	startTime := time.Now()

	// Refresh all shows
	if err := s.crawler.RefreshAll(); err != nil {
		s.logger.Errorf("Weekly crawl failed: %v", err)
	} else {
		s.mu.Lock()
		s.lastCrawlTime = time.Now()
		s.mu.Unlock()
		duration := time.Since(startTime)
		s.logger.Infof("Weekly crawl completed in %v", duration)
	}
}

// weeklyPublishJob performs weekly publish
func (s *Scheduler) weeklyPublishJob() {
	// Check if publish job is already running
	if !s.publishJobMutex.TryLock() {
		s.logger.Warn("Weekly publish job already running, skipping")
		return
	}
	defer s.publishJobMutex.Unlock()

	s.logger.Info("Starting weekly publish job...")
	startTime := time.Now()

	// Publish weekly updates
	result, err := s.publisher.PublishWeeklyUpdates()
	if err != nil {
		s.logger.Errorf("Weekly publish failed: %v", err)
	} else if result.Success {
		s.mu.Lock()
		s.lastPublishTime = time.Now()
		s.mu.Unlock()
		duration := time.Since(startTime)
		s.logger.Infof("Weekly publish completed: %s (%d shows, %d episodes) in %v",
			result.URL,
			result.ShowsCount,
			result.EpisodesCount,
			duration)
	} else {
		s.logger.Warnf("Weekly publish skipped: %v", result.Error)
	}
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
	s.logger.Infof("Immediate crawl completed in %v", duration)
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
		s.logger.Infof("Immediate publish completed: %s (%d shows, %d episodes) in %v",
			result.URL,
			result.ShowsCount,
			result.EpisodesCount,
			duration)
	}

	return result, nil
}

// dailyCorrectionJob performs daily stale show detection
func (s *Scheduler) dailyCorrectionJob() {
	// Check if job is already running
	if !s.correctionJobMutex.TryLock() {
		s.logger.Warn("Daily correction job already running, skipping")
		return
	}
	defer s.correctionJobMutex.Unlock()

	s.logger.Info("Starting daily correction job...")
	startTime := time.Now()

	result, err := s.correction.RunDetection()
	if err != nil {
		s.logger.Errorf("Daily correction job failed: %v", err)
	} else {
		duration := time.Since(startTime)
		s.logger.Infof("Daily correction job completed: %d stale shows found, %d tasks created in %v",
			result.StaleShowsFound, result.TasksCreated, duration)
	}
}

// GetStatus returns the scheduler status
func (s *Scheduler) GetStatus() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := map[string]interface{}{
		"running":              s.running,
		"last_crawl_time":      s.lastCrawlTime,
		"last_publish_time":    s.lastPublishTime,
		"crawl_job_running":    s.crawlJobRunning,
		"publish_job_running":  s.publishJobRunning,
		"correction_job_running": s.correctionJobRunning,
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

// SetTimeouts sets the timeout for crawl and publish jobs
func (s *Scheduler) SetTimeouts(crawlTimeout, publishTimeout time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.crawlTimeout = crawlTimeout
	s.publishTimeout = publishTimeout
}

// GetTimeouts returns the current timeout settings
func (s *Scheduler) GetTimeouts() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return map[string]string{
		"crawl_timeout":   s.crawlTimeout.String(),
		"publish_timeout": s.publishTimeout.String(),
	}
}

// RunManualCrawl runs a manual crawl task
func (s *Scheduler) RunManualCrawl(showID int) error {
	s.logger.Infof("Running manual crawl for show %d", showID)

	if err := s.crawler.CrawlShow(showID); err != nil {
		return fmt.Errorf("manual crawl failed: %w", err)
	}

	s.logger.Infof("Manual crawl completed for show %d", showID)
	return nil
}

// RunManualPublish runs a manual publish task
func (s *Scheduler) RunManualPublish(showID uint) (*PublishResult, error) {
	s.logger.Infof("Running manual publish for show %d", showID)

	result, err := s.publisher.PublishShow(showID)
	if err != nil {
		return nil, fmt.Errorf("manual publish failed: %w", err)
	}

	if result.Success {
		s.mu.Lock()
		s.lastPublishTime = time.Now()
		s.mu.Unlock()

		s.logger.Infof("Manual publish completed: %s", result.URL)
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
// Uses cron.Parse to support 6-field cron expressions with seconds
// Format: seconds minutes hours day month weekday
func ValidateCronSpec(spec string) error {
	// Use cron.Parse instead of cron.ParseStandard to support seconds field
	// This matches the cron.WithSeconds() option used in NewScheduler
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	_, err := parser.Parse(spec)
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

// Helper function to run job with timeout
func (s *Scheduler) runJobWithTimeout(jobName string, timeout time.Duration, job func() error) error {
	s.logger.Infof("Starting %s (timeout: %v)", jobName, timeout)
	startTime := time.Now()

	// Create a channel to receive job result
	done := make(chan error, 1)

	// Run job in goroutine
	go func() {
		done <- job()
	}()

	// Wait for job completion or timeout
	select {
	case err := <-done:
		duration := time.Since(startTime)
		if err != nil {
			s.logger.Errorf("%s failed after %v: %v", jobName, duration, err)
			return err
		}
		s.logger.Infof("%s completed in %v", jobName, duration)
		return nil
	case <-time.After(timeout):
		duration := time.Since(startTime)
		s.logger.Errorf("%s timed out after %v (limit: %v)", jobName, duration, timeout)
		return fmt.Errorf("%s timed out after %v", jobName, timeout)
	}
}

// Helper function to run job with error handling
func (s *Scheduler) runJobWithErrorHandling(jobName string, job func() error) {
	s.logger.Infof("Starting %s", jobName)
	startTime := time.Now()

	if err := job(); err != nil {
		s.logger.Errorf("%s failed: %v", jobName, err)
	} else {
		duration := time.Since(startTime)
		s.logger.Infof("%s completed in %v", jobName, duration)
	}
}
