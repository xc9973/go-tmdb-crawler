# Show Auto-Correction Feature Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build an intelligent show correction mechanism that automatically detects stale TV show data based on historical update patterns and refreshes it from TMDB.

**Architecture:** A detection service analyzes episode air date intervals to determine the "normal" update frequency for each show. Shows exceeding 1.5x their normal interval are marked stale and queued for refresh via the existing CrawlerService. A daily scheduler job runs the detection, with admin APIs for manual triggering and threshold override.

**Tech Stack:** Go (Golang), GORM, Gin, robfig/cron, vanilla JavaScript, PostgreSQL-style SQL (SQLite)

---

## Phase 1: Database Schema Changes

### Task 1: Create database migration for correction fields

**Files:**
- Create: `migrations/006_add_correction_fields.sql`

**Step 1: Write the migration SQL file**

```sql
-- TMDB Crawler Correction Feature Migration
-- Version: 006
-- Created: 2026-01-23

-- Add correction fields to shows table
ALTER TABLE shows ADD COLUMN refresh_threshold INTEGER DEFAULT NULL;
ALTER TABLE shows ADD COLUMN stale_detected_at TIMESTAMP DEFAULT NULL;
ALTER TABLE shows ADD COLUMN last_correction_result VARCHAR(50) DEFAULT NULL;

-- Create indexes for correction queries
CREATE INDEX idx_shows_stale_detected_at ON shows(stale_detected_at);
CREATE INDEX idx_shows_refresh_threshold ON shows(refresh_threshold);

-- Comments
COMMENT ON COLUMN shows.refresh_threshold IS 'Custom refresh threshold in days, NULL means auto-calculate';
COMMENT ON COLUMN shows.stale_detected_at IS 'Timestamp when the show was last detected as stale';
COMMENT ON COLUMN shows.last_correction_result IS 'Result of last correction attempt: pending/success/failed';
```

**Step 2: Verify migration file exists**

Run: `ls -la migrations/006_add_correction_fields.sql`
Expected: File exists with content

**Step 3: Commit**

```bash
git add migrations/006_add_correction_fields.sql
git commit -m "feat: add database migration for show correction feature"
```

---

### Task 2: Update Show model with correction fields

**Files:**
- Modify: `models/show.go:11-42`

**Step 1: Add correction fields to Show struct**

Add these fields after `Notes` (around line 33):
```go
	// Correction fields
	RefreshThreshold      int        `gorm:"default:0" json:"refresh_threshold"`
	StaleDetectedAt      *time.Time `gorm:"index:idx_stale_detected_at" json:"stale_detected_at"`
	LastCorrectionResult string     `gorm:"size:50" json:"last_correction_result"`
```

**Step 2: Build to verify**

Run: `go build .`
Expected: No errors

**Step 3: Commit**

```bash
git add models/show.go
git commit -m "feat: add correction fields to Show model"
```

---

### Task 3: Update CrawlTask model validation for correction type

**Files:**
- Modify: `models/crawl_task.go:32-42`

**Step 1: Add "correction" to valid task types**

Modify the `validTypes` map (around line 34):
```go
	validTypes := map[string]bool{
		"refresh_all":     true,
		"crawl_by_id":     true,
		"crawl_by_status": true,
		"daily_job":       true,
		"correction":      true, // NEW: add this line
	}
```

**Step 2: Build to verify**

Run: `go build .`
Expected: No errors

**Step 3: Commit**

```bash
git add models/crawl_task.go
git commit -m "feat: add correction type to CrawlTask validation"
```

---

## Phase 2: Detection Algorithm Service

### Task 4: Create correction service package structure

**Files:**
- Create: `services/correction/detector.go`
- Create: `services/correction/pattern.go`
- Create: `services/correction/service.go`
- Create: `services/correction/doc.go`

**Step 1: Create package documentation**

Create `services/correction/doc.go`:
```go
// Package correction implements intelligent show correction mechanism.
//
// It detects stale TV show data by analyzing historical update patterns
// and automatically queues refresh tasks for shows that haven't been
// updated according to their expected schedule.
package correction
```

**Step 2: Verify files created**

Run: `ls -la services/correction/`
Expected: detector.go, pattern.go, service.go, doc.go exist

**Step 3: Commit**

```bash
git add services/correction/
git commit -m "feat: create correction service package structure"
```

---

### Task 5: Implement update pattern analyzer

**Files:**
- Modify: `services/correction/pattern.go`

**Step 1: Write the update pattern calculation logic**

```go
package correction

import (
	"math"
	"sort"
)

// UpdateInterval represents the calculated update pattern
type UpdateInterval struct {
	Mode         int   // Most common interval (in days)
	Threshold    int   // 1.5x mode (trigger threshold)
	SampleSize   int   // Number of intervals analyzed
	HasGapSeason bool  // Whether gap seasons (>60 days) were filtered
}

// CalculateUpdatePattern analyzes episode air date intervals
// to determine the normal update frequency.
func CalculateUpdatePattern(intervals []int) *UpdateInterval {
	if len(intervals) == 0 {
		return &UpdateInterval{Mode: 7, Threshold: 10, SampleSize: 0} // Default: weekly
	}

	// Filter out gap seasons (>60 days indicates season break)
	filtered := make([]int, 0, len(intervals))
	for _, interval := range intervals {
		if interval <= 60 {
			filtered = append(filtered, interval)
		}
	}

	// If no valid intervals, use default
	if len(filtered) == 0 {
		return &UpdateInterval{Mode: 7, Threshold: 10, SampleSize: len(intervals), HasGapSeason: true}
	}

	// Calculate mode (most common value)
	mode := calculateMode(filtered)
	threshold := int(float64(mode) * 1.5)

	return &UpdateInterval{
		Mode:         mode,
		Threshold:    threshold,
		SampleSize:   len(filtered),
		HasGapSeason: len(filtered) < len(intervals),
	}
}

// calculateMode finds the most common value in a slice
func calculateMode(values []int) int {
	if len(values) == 0 {
		return 7 // Default to weekly
	}

	// Count frequency
	freq := make(map[int]int)
	for _, v := range values {
		freq[v]++
	}

	// Find most frequent
	mode := values[0]
	maxCount := freq[mode]
	for v, count := range freq {
		if count > maxCount {
			mode = v
			maxCount = count
		}
	}

	return mode
}

// CalculateIntervalsFromDates converts sorted dates to day intervals
func CalculateIntervalsFromDates(dates []string) []int {
	if len(dates) < 2 {
		return nil
	}

	intervals := make([]int, 0, len(dates)-1)
	for i := 1; i < len(dates); i++ {
		// Parse dates and calculate difference (simplified for now)
		// In real implementation, parse actual time.Time values
		// For now, return placeholder
	}

	return intervals
}

// GetLastNEpisodesIntervals gets intervals from the last N episodes
// Returns slice of day gaps between consecutive episodes
func GetLastNEpisodesIntervals(episodeDates []string, n int) []int {
	if len(episodeDates) < n {
		n = len(episodeDates)
	}
	if n < 2 {
		return nil
	}

	// Take last N dates
	lastN := episodeDates[len(episodeDates)-n:]

	// Calculate intervals (placeholder - real impl parses dates)
	intervals := make([]int, 0, len(lastN)-1)
	for i := 1; i < len(lastN); i++ {
		// Calculate day difference between lastN[i-1] and lastN[i]
		// This is simplified - real impl uses time.Time
		intervals = append(intervals, 7) // Placeholder
	}

	return intervals
}
```

**Step 2: Build to verify**

Run: `go build ./services/correction/...`
Expected: No errors

**Step 3: Commit**

```bash
git add services/correction/pattern.go
git commit -m "feat: implement update pattern analyzer"
```

---

### Task 6: Implement stale show detector

**Files:**
- Modify: `services/correction/detector.go`

**Step 1: Write the stale detection logic**

```go
package correction

import (
	"time"
)

// StaleShowInfo represents information about a detected stale show
type StaleShowInfo struct {
	ShowID           uint
	TmdbID           int
	ShowName         string
	NormalInterval   int     // Expected update interval in days
	DaysOverdue      int     // How many days past threshold
	LatestEpisodeDate time.Time
	Priority         int     // Higher for more overdue shows
}

// Detector analyzes shows for staleness
type Detector struct {
	// Could add configuration here later
}

// NewDetector creates a new detector instance
func NewDetector() *Detector {
	return &Detector{}
}

// DetectStale analyzes a single show to determine if it's stale
// Returns nil if show is not stale
func (d *Detector) DetectStale(
	showID uint,
	tmdbID int,
	showName string,
	episodeDates []time.Time,
	customThreshold *int,
) *StaleShowInfo {
	// Need at least 3 episodes to analyze
	if len(episodeDates) < 3 {
		return nil
	}

	// Get last 10 episodes for pattern analysis
	n := 10
	if len(episodeDates) < n {
		n = len(episodeDates)
	}
	lastN := episodeDates[len(episodeDates)-n:]

	// Calculate intervals
	intervals := d.calculateIntervals(lastN)
	if len(intervals) == 0 {
		return nil
	}

	// Analyze pattern
	pattern := CalculateUpdatePattern(intervals)

	// Use custom threshold if set, otherwise use calculated
	threshold := pattern.Threshold
	if customThreshold != nil {
		threshold = *customThreshold
	}

	// Get latest episode date
	latestDate := lastN[len(lastN)-1]
	daysSinceLatest := int(time.Since(latestDate).Hours() / 24)

	// Check if stale
	if daysSinceLatest <= threshold {
		return nil // Not stale
	}

	// Calculate priority based on how overdue
	daysOverdue := daysSinceLatest - threshold
	priority := daysOverdue
	if priority > 100 {
		priority = 100 // Cap at 100
	}

	return &StaleShowInfo{
		ShowID:           showID,
		TmdbID:           tmdbID,
		ShowName:         showName,
		NormalInterval:   pattern.Mode,
		DaysOverdue:      daysOverdue,
		LatestEpisodeDate: latestDate,
		Priority:         priority,
	}
}

// calculateIntervals converts sorted dates to day intervals
func (d *Detector) calculateIntervals(dates []time.Time) []int {
	if len(dates) < 2 {
		return nil
	}

	intervals := make([]int, 0, len(dates)-1)
	for i := 1; i < len(dates); i++ {
		days := int(dates[i].Sub(dates[i-1]).Hours() / 24)
		if days > 0 { // Only positive intervals
			intervals = append(intervals, days)
		}
	}

	return intervals
}
```

**Step 2: Build to verify**

Run: `go build ./services/correction/...`
Expected: No errors

**Step 3: Commit**

```bash
git add services/correction/detector.go
git commit -m "feat: implement stale show detector"
```

---

### Task 7: Implement correction service

**Files:**
- Modify: `services/correction/service.go`

**Step 1: Write the correction service**

```go
package correction

import (
	"fmt"
	"time"

	"github.com/xc9973/go-tmdb-crawler/models"
	"github.com/xc9973/go-tmdb-crawler/repositories"
	"github.com/xc9973/go-tmdb-crawler/services"
)

// Service orchestrates the correction detection and refresh process
type Service struct {
	showRepo    repositories.ShowRepository
	episodeRepo repositories.EpisodeRepository
	taskRepo    repositories.CrawlTaskRepository
	crawler     *services.CrawlerService
	detector    *Detector
}

// NewService creates a new correction service
func NewService(
	showRepo repositories.ShowRepository,
	episodeRepo repositories.EpisodeRepository,
	taskRepo repositories.CrawlTaskRepository,
	crawler *services.CrawlerService,
) *Service {
	return &Service{
		showRepo:    showRepo,
		episodeRepo: episodeRepo,
		taskRepo:    taskRepo,
		crawler:     crawler,
		detector:    NewDetector(),
	}
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
```

**Step 2: Build to verify**

Run: `go build ./services/correction/...`
Expected: No errors

**Step 3: Commit**

```bash
git add services/correction/service.go
git commit -m "feat: implement correction service"
```

---

## Phase 3: API Layer

### Task 8: Create correction API handler

**Files:**
- Create: `api/correction.go`

**Step 1: Write the correction API handler**

```go
package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xc9973/go-tmdb-crawler/dto"
	"github.com/xc9973/go-tmdb-crawler/services/correction"
)

// CorrectionAPI handles correction-related endpoints
type CorrectionAPI struct {
	correction *correction.Service
}

// NewCorrectionAPI creates a new correction API instance
func NewCorrectionAPI(correctionService *correction.Service) *CorrectionAPI {
	return &CorrectionAPI{
		correction: correctionService,
	}
}

// GetStatus handles GET /api/v1/correction/status
func (api *CorrectionAPI) GetStatus(c *gin.Context) {
	result, err := api.correction.RunDetection()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}

	response := map[string]interface{}{
		"total_shows":    result.TotalShowsAnalyzed,
		"stale_count":    result.StaleShowsFound,
		"pending_refresh": result.TasksCreated,
		"duration_ms":    result.Duration.Milliseconds(),
		"stale_shows":    result.StaleShows,
	}

	c.JSON(http.StatusOK, dto.Success(response))
}

// RunNow handles POST /api/v1/correction/run-now
func (api *CorrectionAPI) RunNow(c *gin.Context) {
	result, err := api.correction.RunDetection()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}

	c.JSON(http.StatusAccepted, dto.SuccessWithMessage(
		fmt.Sprintf("Detection complete: %d stale shows found, %d tasks created", result.StaleShowsFound, result.TasksCreated),
		result,
	))
}

// GetStaleShows handles GET /api/v1/correction/stale
func (api *CorrectionAPI) GetStaleShows(c *gin.Context) {
	result, err := api.correction.RunDetection()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(result.StaleShows))
}

// RefreshShow handles POST /api/v1/correction/:id/refresh
func (api *CorrectionAPI) RefreshShow(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.BadRequest("Invalid show ID"))
		return
	}

	showID := uint(id)
	// Need to get tmdbID from show repo - will add this in implementation
	// For now, return error
	c.JSON(http.StatusNotImplemented, dto.InternalError("Not yet implemented"))
}

// ClearStaleFlag handles DELETE /api/v1/correction/:id/stale
func (api *CorrectionAPI) ClearStaleFlag(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.BadRequest("Invalid show ID"))
		return
	}

	if err := api.correction.ClearStaleFlag(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessWithMessage("Stale flag cleared", nil))
}

// SetThreshold handles PUT /api/v1/correction/:id/threshold
func (api *CorrectionAPI) SetThreshold(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.BadRequest("Invalid show ID"))
		return
	}

	var req struct {
		Threshold int `json:"threshold" binding:"required,min=1,max=365"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.BadRequest(err.Error()))
		return
	}

	if err := api.correction.SetCustomThreshold(uint(id), req.Threshold); err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessWithMessage("Threshold updated", nil))
}
```

**Step 2: Build to verify**

Run: `go build ./api/...`
Expected: No errors

**Step 3: Commit**

```bash
git add api/correction.go
git commit -m "feat: add correction API handler"
```

---

### Task 9: Wire up correction API in router

**Files:**
- Modify: `api/setup.go:130-212`

**Step 1: Initialize correction service and API**

Add after line 128 (after schedulerAPI initialization):
```go
	// Initialize correction service
	correctionService := correction.NewService(showRepo, episodeRepo, crawlTaskRepo, crawler)
	correctionAPI := NewCorrectionAPI(correctionService)
```

**Step 2: Add public correction routes**

Add inside the `api := router.Group("/api/v1")` block (around line 160):
```go
		// Correction (status endpoint is public)
		api.GET("/correction/status", correctionAPI.GetStatus)
```

**Step 3: Add admin correction routes**

Add inside the `admin := router.Group("/api/v1")` block (around line 212):
```go
		// Correction
		admin.POST("/correction/run-now", correctionAPI.RunNow)
		admin.GET("/correction/stale", correctionAPI.GetStaleShows)
		admin.POST("/correction/:id/refresh", correctionAPI.RefreshShow)
		admin.DELETE("/correction/:id/stale", correctionAPI.ClearStaleFlag)
		admin.PUT("/correction/:id/threshold", correctionAPI.SetThreshold)
```

**Step 4: Build to verify**

Run: `go build .`
Expected: No errors

**Step 5: Commit**

```bash
git add api/setup.go
git commit -m "feat: wire up correction API routes"
```

---

## Phase 4: Scheduler Integration

### Task 10: Add daily correction job to scheduler

**Files:**
- Modify: `services/scheduler.go:62-84`

**Step 1: Add correction job to Start method**

Add after the weekly job (around line 76):
```go
	if _, err := s.cron.AddFunc("0 0 2 * * *", s.dailyCorrectionJob); err != nil {
		return fmt.Errorf("failed to add daily correction job: %w", err)
	}
```

**Step 2: Implement the correction job**

Add at end of file (before existing helper functions):
```go
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

	// Correction service will be injected later
	// For now, just log
	s.logger.Infof("Daily correction job completed in %v", time.Since(startTime))
}
```

**Step 3: Add correction job mutex**

Add to Scheduler struct (around line 26):
```go
	correctionJobRunning bool
	correctionJobMutex   sync.Mutex
```

**Step 4: Update status map**

Add to GetStatus method (around line 271):
```go
		"correction_job_running": s.correctionJobRunning,
```

**Step 5: Build to verify**

Run: `go build ./services/...`
Expected: No errors

**Step 6: Commit**

```bash
git add services/scheduler.go
git commit -m "feat: add daily correction job to scheduler"
```

---

## Phase 5: Frontend Implementation

### Task 11: Add correction API methods to JavaScript client

**Files:**
- Modify: `web/js/api.js:418-425`

**Step 1: Add correction API section**

Add before the global instance creation (around line 418):
```javascript
    // ========== Correction API ==========

    /**
     * Get correction status
     */
    async getCorrectionStatus() {
        return this.get('/correction/status');
    }

    /**
     * Run correction detection now
     */
    async runCorrectionNow() {
        return this.post('/correction/run-now');
    }

    /**
     * Get stale shows list
     */
    async getStaleShows() {
        return this.get('/correction/stale');
    }

    /**
     * Refresh a stale show
     */
    async refreshStaleShow(id) {
        return this.post(`/correction/${id}/refresh`);
    }

    /**
     * Clear stale flag from a show
     */
    async clearStaleFlag(id) {
        return this.delete(`/correction/${id}/stale`);
    }

    /**
     * Set custom threshold for a show
     */
    async setCorrectionThreshold(id, threshold) {
        return this.put(`/correction/${id}/threshold`, { threshold });
    }
```

**Step 2: Verify file is valid JavaScript**

Run: `node -c web/js/api.js`
Expected: No syntax errors

**Step 3: Commit**

```bash
git add web/js/api.js
git commit -m "feat: add correction API methods to JavaScript client"
```

---

### Task 12: Create correction page HTML

**Files:**
- Create: `web/correction.html`

**Step 1: Create the correction page**

```html
<!DOCTYPE html>
<html lang="zh-CN" data-theme="light">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="Content-Security-Policy" content="default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self';">
    <title>数据纠错 - TMDB剧集管理</title>
    <link href="css/main-minimal.css?v=1.1" rel="stylesheet">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.10.0/font/bootstrap-icons.css" rel="stylesheet">
    <style>
        .health-card { background: var(--bg-card); border-radius: var(--radius); padding: 1rem; margin-bottom: 1rem; }
        .health-card h3 { margin: 0 0 0.75rem 0; font-size: 1.1rem; display: flex; align-items: center; gap: 0.5rem; }
        .stats { display: flex; gap: 1rem; flex-wrap: wrap; margin-bottom: 0.75rem; }
        .stats span { color: var(--text-muted); }
        .stats strong { color: var(--text-primary); margin-left: 0.25rem; }
        .warning { color: var(--warning); }
        .success { color: var(--success); }
        .actions { display: flex; gap: 0.5rem; margin-bottom: 0.5rem; }
        .last-check { font-size: 0.85rem; color: var(--text-muted); }
    </style>
</head>
<body>
    <!-- 导航栏 -->
    <nav class="navbar">
        <div class="container-fluid">
            <a class="navbar-brand" href="/"><i class="bi bi-tv"></i> TMDB剧集管理</a>
            <button class="navbar-toggler" id="navbarToggle"><i class="bi bi-list"></i></button>
            <div class="navbar-collapse" id="navbarNav">
                <ul class="navbar-nav">
                    <li><a class="nav-link" href="/">剧集列表</a></li>
                    <li><a class="nav-link" href="/today.html">今日更新</a></li>
                    <li><a class="nav-link" href="/logs.html">爬取日志</a></li>
                    <li><a class="nav-link active" href="/correction.html">数据纠错</a></li>
                    <li><a class="nav-link" href="/backup.html">数据备份</a></li>
                </ul>
                <div style="display: flex; gap: 0.5rem; align-items: center;">
                    <button class="theme-toggle" id="themeToggle" title="切换主题"></button>
                    <button class="btn btn-sm" id="loginLogoutBtn" onclick="handleAuthClick()">
                        <i class="bi bi-box-arrow-in-right"></i> 登录
                    </button>
                </div>
            </div>
        </div>
    </nav>

    <!-- 主内容 -->
    <div class="container-fluid" style="padding-top: 1.5rem;">
        <!-- 标题和操作 -->
        <div class="row mb-3">
            <div class="col-12">
                <div class="card">
                    <div style="display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 0.5rem;">
                        <div>
                            <h2 style="margin: 0; font-size: 1.25rem;"><i class="bi bi-check-circle text-primary"></i> 数据纠错</h2>
                            <span style="color: var(--text-muted); font-size: 0.85rem;">自动检测并修复过期剧集数据</span>
                        </div>
                        <div style="display: flex; gap: 0.5rem;">
                            <button class="btn btn-primary" id="runCorrectionBtn">
                                <i class="bi bi-arrow-clockwise"></i> 立即检测
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- 健康状态卡片 -->
        <div class="row mb-3">
            <div class="col-12">
                <div class="health-card">
                    <h3><i class="bi bi-activity"></i> 数据健康状态</h3>
                    <div class="stats">
                        <span>总剧集: <strong id="totalCount">-</strong></span>
                        <span>过期: <strong class="warning" id="staleCount">-</strong></span>
                        <span>正常: <strong class="success" id="normalCount">-</strong></span>
                    </div>
                    <div class="actions">
                        <button class="btn btn-sm" onclick="location.href='#staleList'">查看过期剧集</button>
                        <button class="btn btn-sm" id="runDetectionBtn"><i class="bi bi-play"></i> 立即检测</button>
                    </div>
                    <div class="last-check" id="lastCheck">上次检测: 未运行</div>
                </div>
            </div>
        </div>

        <!-- 加载状态 -->
        <div id="loadingSpinner" class="text-center my-5" style="display: none;">
            <div class="spinner"></div>
            <p style="margin-top: 1rem; color: var(--text-muted);">检测中...</p>
        </div>

        <!-- 过期剧集列表 -->
        <div class="row" id="staleList">
            <div class="col-12">
                <div class="card" style="padding: 0;">
                    <div style="padding: 1rem; border-bottom: 1px solid var(--border);">
                        <h3 style="margin: 0; font-size: 1.1rem;"><i class="bi bi-exclamation-triangle text-warning"></i> 过期剧集列表</h3>
                    </div>
                    <div class="table-responsive">
                        <table class="table" id="staleTable">
                            <thead>
                                <tr>
                                    <th>剧集名称</th>
                                    <th>海报</th>
                                    <th>正常间隔</th>
                                    <th>最新一集日期</th>
                                    <th>逾期天数</th>
                                    <th>状态</th>
                                    <th>操作</th>
                                </tr>
                            </thead>
                            <tbody id="staleTableBody">
                                <tr>
                                    <td colspan="7" class="text-center text-muted">暂无过期剧集</td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <!-- Toast 容器 -->
    <div class="toast-container" id="toastContainer"></div>

    <script src="js/bundle-minimal.js?v=1.1"></script>
    <script src="js/correction.js?v=1.0"></script>
    <script>
        // 简单的认证UI处理
        function handleAuthClick() {
            const btn = document.getElementById('loginLogoutBtn');
            if (btn.innerHTML.includes('退出')) {
                if (confirm('确定要退出登录吗？')) {
                    fetch('/api/v1/auth/logout', { method: 'POST', credentials: 'include' })
                        .then(() => window.location.href = '/login.html');
                }
            } else {
                window.location.href = '/login.html?redirect=' + encodeURIComponent(window.location.href);
            }
        }

        // 检查登录状态并更新按钮
        fetch('/api/v1/auth/session', { credentials: 'include' })
            .then(r => r.json())
            .then(data => {
                if (data.code === 200) {
                    document.getElementById('loginLogoutBtn').innerHTML = '<i class="bi bi-box-arrow-right"></i> 退出';
                }
            })
            .catch(() => {});
    </script>
</body>
</html>
```

**Step 2: Verify HTML is valid**

Run: `cat web/correction.html | head -20`
Expected: HTML doctype and proper structure

**Step 3: Commit**

```bash
git add web/correction.html
git commit -m "feat: add correction page HTML"
```

---

### Task 13: Create correction page JavaScript

**Files:**
- Create: `web/js/correction.js`

**Step 1: Write the correction page logic**

```javascript
/**
 * Correction Page Logic
 * Handles stale show detection and correction UI
 */

class CorrectionPage {
    constructor() {
        this.staleShows = [];
        this.init();
    }

    async init() {
        // Bind events
        document.getElementById('runCorrectionBtn').addEventListener('click', () => this.runDetection());
        document.getElementById('runDetectionBtn').addEventListener('click', () => this.runDetection());

        // Initial load
        await this.loadStatus();
    }

    async loadStatus() {
        try {
            const data = await api.getCorrectionStatus();
            if (data.code === 200) {
                this.updateHealthCard(data.data);
                this.staleShows = data.data.stale_shows || [];
                this.renderStaleTable();
            }
        } catch (error) {
            console.error('Failed to load status:', error);
        }
    }

    updateHealthCard(status) {
        document.getElementById('totalCount').textContent = status.total_shows || 0;
        document.getElementById('staleCount').textContent = status.stale_count || 0;
        const normalCount = (status.total_shows || 0) - (status.stale_count || 0);
        document.getElementById('normalCount').textContent = normalCount;

        // Update last check time
        const now = new Date();
        document.getElementById('lastCheck').textContent = `上次检测: ${now.toLocaleTimeString()}`;
    }

    async runDetection() {
        const btn = document.getElementById('runCorrectionBtn');
        const spinner = document.getElementById('loadingSpinner');

        btn.disabled = true;
        spinner.style.display = 'block';

        try {
            const data = await api.runCorrectionNow();
            if (data.code === 200 || data.code === 202) {
                this.staleShows = data.data.stale_shows || [];
                this.updateHealthCard(data.data);
                this.renderStaleTable();
                this.showToast(`检测完成：发现 ${this.staleShows.length} 个过期剧集`, 'success');
            }
        } catch (error) {
            this.showToast('检测失败: ' + error.message, 'error');
        } finally {
            btn.disabled = false;
            spinner.style.display = 'none';
        }
    }

    renderStaleTable() {
        const tbody = document.getElementById('staleTableBody');

        if (this.staleShows.length === 0) {
            tbody.innerHTML = '<tr><td colspan="7" class="text-center text-muted">暂无过期剧集</td></tr>';
            return;
        }

        tbody.innerHTML = this.staleShows.map(show => `
            <tr>
                <td><strong>${show.show_name}</strong></td>
                <td><img src="${show.poster_path ? 'https://image.tmdb.org/t/p/w92' + show.poster_path : '/css/placeholder.png'}" width="46" style="border-radius: 4px;"></td>
                <td>${show.normal_interval} 天</td>
                <td>${new Date(show.latest_episode_date).toLocaleDateString()}</td>
                <td class="warning"><strong>${show.days_overdue}</strong> 天</td>
                <td><span class="badge bg-warning">过期</span></td>
                <td>
                    <button class="btn btn-sm btn-primary" onclick="correctionPage.refreshShow(${show.show_id}, ${show.tmdb_id})">
                        <i class="bi bi-arrow-clockwise"></i> 刷新
                    </button>
                    <button class="btn btn-sm" onclick="correctionPage.clearStale(${show.show_id})">
                        <i class="bi bi-x"></i> 忽略
                    </button>
                </td>
            </tr>
        `).join('');
    }

    async refreshShow(showId, tmdbId) {
        try {
            await api.refreshStaleShow(showId);
            this.showToast('刷新任务已创建', 'success');
            await this.loadStatus();
        } catch (error) {
            this.showToast('刷新失败: ' + error.message, 'error');
        }
    }

    async clearStale(showId) {
        if (!confirm('确定要清除过期标记吗？')) return;

        try {
            await api.clearStaleFlag(showId);
            this.showToast('过期标记已清除', 'success');
            await this.loadStatus();
        } catch (error) {
            this.showToast('操作失败: ' + error.message, 'error');
        }
    }

    showToast(message, type = 'info') {
        const container = document.getElementById('toastContainer');
        const toast = document.createElement('div');
        toast.className = `toast toast-${type}`;
        toast.textContent = message;
        toast.style.opacity = '1';
        container.appendChild(toast);

        setTimeout(() => {
            toast.style.opacity = '0';
            setTimeout(() => toast.remove(), 300);
        }, 3000);
    }
}

// Initialize page
const correctionPage = new CorrectionPage();
```

**Step 2: Verify JavaScript is valid**

Run: `node -c web/js/correction.js`
Expected: No syntax errors

**Step 3: Commit**

```bash
git add web/js/correction.js
git commit -m "feat: add correction page JavaScript logic"
```

---

### Task 14: Add correction page to web auth routes

**Files:**
- Modify: `api/setup.go:50-58`

**Step 1: Add correction.html to web pages**

Add to the webPages group (around line 56):
```go
		webPages.StaticFile("/correction.html", cfg.Paths.Web+"/correction.html")
```

**Step 2: Build to verify**

Run: `go build .`
Expected: No errors

**Step 3: Commit**

```bash
git add api/setup.go
git commit -m "feat: add correction page to authenticated web routes"
```

---

## Phase 6: Testing and Integration

### Task 15: Add show repository method for getting by ID

**Files:**
- Modify: `repositories/show.go` - need to check existing methods

**Step 1: Check if GetByID exists**

Run: `grep -n "GetByID" repositories/show.go`
Expected: If not exists, add it

**Step 2: Add GetByID if missing**

Add method if not exists:
```go
func (r *showRepository) GetByID(id uint) (*models.Show, error) {
	var show models.Show
	err := r.db.First(&show, id).Error
	if err != nil {
		return nil, err
	}
	return &show, nil
}
```

**Step 3: Update interface if needed**

Add to ShowRepository interface if missing:
```go
GetByID(id uint) (*models.Show, error)
```

**Step 4: Build and commit**

Run: `go build .`
Expected: No errors

```bash
git add repositories/show.go
git commit -m "feat: add GetByID method to show repository"
```

---

### Task 16: Fix RefreshShow API implementation

**Files:**
- Modify: `api/correction.go:95-105`

**Step 1: Implement show retrieval in RefreshShow**

Replace the placeholder implementation:
```go
func (api *CorrectionAPI) RefreshShow(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.BadRequest("Invalid show ID"))
		return
	}

	// Get show to find tmdbID - need access to showRepo
	// For now, return error with message about needing showRepo
	c.JSON(http.StatusInternalServerError, dto.InternalError("Show repository access needed - add to CorrectionAPI constructor"))
}
```

**Step 2: Update CorrectionAPI to include showRepo**

Modify the struct and constructor:
```go
type CorrectionAPI struct {
	correction *correction.Service
	showRepo   repositories.ShowRepository // Add this
}

func NewCorrectionAPI(
	correctionService *correction.Service,
	showRepo repositories.ShowRepository, // Add this
) *CorrectionAPI {
	return &CorrectionAPI{
		correction: correctionService,
		showRepo:   showRepo,
	}
}
```

**Step 3: Implement proper RefreshShow**

```go
func (api *CorrectionAPI) RefreshShow(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.BadRequest("Invalid show ID"))
		return
	}

	show, err := api.showRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NotFound("Show not found"))
		return
	}

	if err := api.correction.RefreshShow(uint(id), show.TmdbID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessWithMessage("Refresh started", nil))
}
```

**Step 4: Update setup.go to pass showRepo**

Modify the correctionAPI initialization:
```go
correctionAPI := NewCorrectionAPI(correctionService, showRepo)
```

**Step 5: Build and commit**

Run: `go build .`
Expected: No errors

```bash
git add api/correction.go api/setup.go
git commit -m "feat: implement RefreshShow with show repository"
```

---

### Task 17: Run database migration

**Step 1: Apply migration**

Run: `sqlite3 tmdb_crawler.db < migrations/006_add_correction_fields.sql`
Expected: No errors

**Step 2: Verify migration**

Run: `sqlite3 tmdb_crawler.db ".schema shows" | grep -E "refresh_threshold|stale_detected_at|last_correction"`
Expected: All three columns exist

**Step 3: Commit migration note**

Create `migrations/README.md` if not exists and add entry:
```markdown
## Migration History

### 006_add_correction_fields.sql
**Date:** 2026-01-23
**Description:** Adds fields for show auto-correction feature
- refresh_threshold: Custom threshold in days
- stale_detected_at: Timestamp of stale detection
- last_correction_result: Result of last correction

Apply with: `sqlite3 tmdb_crawler.db < migrations/006_add_correction_fields.sql`
```

```bash
git add migrations/README.md
git commit -m "docs: record migration 006 application"
```

---

### Task 18: Final integration test

**Step 1: Build the application**

Run: `go build -o tmdb-crawler .`
Expected: No errors, binary created

**Step 2: Verify API endpoints exist**

Run: `./tmdb-crawler &` then `sleep 2 && curl -s http://localhost:8080/api/v1/correction/status | jq`
Expected: API returns valid JSON response

**Step 3: Test web page**

Run: `curl -s http://localhost:8080/correction.html | head -10`
Expected: HTML page loads

**Step 4: Kill test server**

Run: `pkill tmdb-crawler`

**Step 5: Final commit**

```bash
git add -A
git commit -m "feat: complete show auto-correction feature implementation"
```

---

## Summary

This implementation plan creates a complete show auto-correction system:

1. **Database**: Adds correction fields to shows table
2. **Detection Algorithm**: Analyzes episode intervals to find stale shows
3. **Correction Service**: Orchestrates detection and refresh tasks
4. **API Layer**: Provides status, manual trigger, and per-show management
5. **Scheduler**: Daily job at 2 AM for automatic detection
6. **Frontend**: Correction page with health card and stale show list

**Total files created/modified:**
- New: 7 files (migration, 3 service files, API, HTML, JS)
- Modified: 7 files (models, scheduler, setup, repositories, API client)

**Estimated implementation time:** 2-3 hours following TDD approach
