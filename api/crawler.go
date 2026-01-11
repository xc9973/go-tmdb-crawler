package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xc9973/go-tmdb-crawler/dto"
	"github.com/xc9973/go-tmdb-crawler/models"
	"github.com/xc9973/go-tmdb-crawler/repositories"
	"github.com/xc9973/go-tmdb-crawler/services"
)

// CrawlerAPI handles crawler control endpoints
type CrawlerAPI struct {
	crawler     *services.CrawlerService
	showRepo    repositories.ShowRepository
	logRepo     repositories.CrawlLogRepository
	episodeRepo repositories.EpisodeRepository
}

// NewCrawlerAPI creates a new crawler API instance
func NewCrawlerAPI(
	crawler *services.CrawlerService,
	showRepo repositories.ShowRepository,
	logRepo repositories.CrawlLogRepository,
	episodeRepo repositories.EpisodeRepository,
) *CrawlerAPI {
	return &CrawlerAPI{
		crawler:     crawler,
		showRepo:    showRepo,
		logRepo:     logRepo,
		episodeRepo: episodeRepo,
	}
}

// RefreshAll handles POST /api/v1/crawler/refresh-all
func (api *CrawlerAPI) RefreshAll(c *gin.Context) {
	// Start refresh in background
	go func() {
		_ = api.crawler.RefreshAll()
	}()

	c.JSON(http.StatusAccepted, dto.SuccessWithMessage("Refresh started in background", nil))
}

// GetTodayUpdates handles GET /api/v1/crawler/today-updates
func (api *CrawlerAPI) GetTodayUpdates(c *gin.Context) {
	// Get episodes airing today
	episodes, err := api.episodeRepo.GetTodayUpdates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}

	// Build response with show information
	type EpisodeWithShow struct {
		*models.Episode
		ShowName string `json:"show_name"`
	}

	result := make([]EpisodeWithShow, 0)
	for _, episode := range episodes {
		item := EpisodeWithShow{
			Episode:  episode,
			ShowName: episode.Show.Name,
		}
		result = append(result, item)
	}

	c.JSON(http.StatusOK, dto.Success(result))
}

// GetCrawlLogs handles GET /api/v1/crawler/logs
func (api *CrawlerAPI) GetCrawlLogs(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")

	// Validate
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var logs []*models.CrawlLog
	var total int64
	var err error

	// Filter by status if provided
	if status != "" {
		logs, total, err = api.logRepo.GetByStatus(status, page, pageSize)
	} else {
		logs, err = api.logRepo.GetRecent(pageSize)
		if err == nil {
			total = int64(len(logs))
		}
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}

	// Build response
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	response := dto.ListResponse{
		Items:      logs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, dto.Success(response))
}

// GetCrawlStats handles GET /api/v1/crawler/stats
func (api *CrawlerAPI) GetCrawlStats(c *gin.Context) {
	// Get statistics
	showCount, err := api.showRepo.Count()
	if err != nil {
		showCount = 0
	}

	episodeCount, err := api.episodeRepo.Count()
	if err != nil {
		episodeCount = 0
	}

	successCount, err := api.logRepo.CountByStatus("success")
	if err != nil {
		successCount = 0
	}

	failedCount, err := api.logRepo.CountByStatus("failed")
	if err != nil {
		failedCount = 0
	}

	todayEpisodes, err := api.episodeRepo.GetTodayUpdates()
	if err != nil {
		todayEpisodes = nil
	}

	stats := map[string]interface{}{
		"total_shows":       showCount,
		"total_episodes":    episodeCount,
		"successful_crawls": successCount,
		"failed_crawls":     failedCount,
		"today_updates":     len(todayEpisodes),
	}

	c.JSON(http.StatusOK, dto.Success(stats))
}

// GetUpdatesByDateRange handles GET /api/v1/crawler/updates
func (api *CrawlerAPI) GetUpdatesByDateRange(c *gin.Context) {
	// Parse query parameters
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, dto.BadRequest("start_date and end_date are required"))
		return
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.BadRequest("Invalid start_date format. Use YYYY-MM-DD"))
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.BadRequest("Invalid end_date format. Use YYYY-MM-DD"))
		return
	}

	// Get episodes in date range
	episodes, err := api.episodeRepo.GetByDateRange(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}

	// Build response with show information
	type EpisodeWithShow struct {
		*models.Episode
		ShowName string `json:"show_name"`
	}

	result := make([]EpisodeWithShow, 0)
	for _, episode := range episodes {
		item := EpisodeWithShow{
			Episode:  episode,
			ShowName: episode.Show.Name,
		}
		result = append(result, item)
	}

	c.JSON(http.StatusOK, dto.Success(result))
}

// CrawlByStatus handles POST /api/v1/crawler/crawl-by-status
func (api *CrawlerAPI) CrawlByStatus(c *gin.Context) {
	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.BadRequest(err.Error()))
		return
	}

	// Get shows by status
	var shows []*models.Show
	var err error

	if req.Status == "returning" || req.Status == "Returning Series" {
		shows, err = api.showRepo.ListReturning()
	} else {
		shows, err = api.showRepo.ListAll()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}

	// Extract TMDB IDs
	tmdbIDs := make([]int, len(shows))
	for i, show := range shows {
		tmdbIDs[i] = show.TmdbID
	}

	// Batch crawl
	go func() {
		_ = api.crawler.BatchCrawl(tmdbIDs)
	}()

	c.JSON(http.StatusAccepted, dto.SuccessWithMessage(
		fmt.Sprintf("Started crawling %d shows with status: %s", len(shows), req.Status),
		nil,
	))
}

// DeleteOldLogs handles DELETE /api/v1/crawler/logs/old
func (api *CrawlerAPI) DeleteOldLogs(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.BadRequest("Invalid days parameter"))
		return
	}

	if err := api.logRepo.DeleteOld(days); err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessWithMessage(
		fmt.Sprintf("Deleted crawl logs older than %d days", days),
		nil,
	))
}

// GetHealthStatus handles GET /api/v1/crawler/health
func (api *CrawlerAPI) GetHealthStatus(c *gin.Context) {
	// Check if recent crawls are successful
	recentLogs, err := api.logRepo.GetRecent(10)
	if err != nil {
		c.JSON(http.StatusOK, dto.Success(map[string]interface{}{
			"status":  "unknown",
			"message": "Unable to retrieve recent logs",
		}))
		return
	}

	// Calculate success rate
	successCount := 0
	for _, log := range recentLogs {
		if log.IsSuccess() {
			successCount++
		}
	}

	successRate := float64(0)
	if len(recentLogs) > 0 {
		successRate = float64(successCount) / float64(len(recentLogs)) * 100
	}

	status := "healthy"
	if successRate < 50 {
		status = "unhealthy"
	} else if successRate < 80 {
		status = "degraded"
	}

	c.JSON(http.StatusOK, dto.Success(map[string]interface{}{
		"status":        status,
		"success_rate":  fmt.Sprintf("%.1f%%", successRate),
		"recent_crawls": len(recentLogs),
	}))
}
