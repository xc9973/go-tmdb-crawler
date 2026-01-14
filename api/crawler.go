package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xc9973/go-tmdb-crawler/dto"
	"github.com/xc9973/go-tmdb-crawler/models"
	"github.com/xc9973/go-tmdb-crawler/repositories"
	"github.com/xc9973/go-tmdb-crawler/services"
	"gorm.io/gorm"
)

// CrawlerAPI handles crawler control endpoints
type CrawlerAPI struct {
	crawler     *services.CrawlerService
	showRepo    repositories.ShowRepository
	logRepo     repositories.CrawlLogRepository
	episodeRepo repositories.EpisodeRepository
	taskManager *services.TaskManager
}

// NewCrawlerAPI creates a new crawler API instance
func NewCrawlerAPI(
	crawler *services.CrawlerService,
	showRepo repositories.ShowRepository,
	logRepo repositories.CrawlLogRepository,
	episodeRepo repositories.EpisodeRepository,
	taskManager *services.TaskManager,
) *CrawlerAPI {
	return &CrawlerAPI{
		crawler:     crawler,
		showRepo:    showRepo,
		logRepo:     logRepo,
		episodeRepo: episodeRepo,
		taskManager: taskManager,
	}
}

// SearchTMDB handles GET /api/v1/crawler/search/tmdb
func (api *CrawlerAPI) SearchTMDB(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, dto.BadRequest("query parameter is required"))
		return
	}

	// 如果query是纯数字，视为TMDB ID搜索
	var tmdbID int
	if _, err := fmt.Sscanf(query, "%d", &tmdbID); err == nil {
		// 通过TMDB ID获取详情
		tmdbShow, err := api.crawler.GetTMDBService().GetShowDetails(tmdbID)
		if err != nil {
			c.JSON(http.StatusNotFound, dto.InternalError("TMDB搜索失败: "+err.Error()))
			return
		}

		// 转换为搜索结果格式
		result := map[string]interface{}{
			"page": 1,
			"results": []map[string]interface{}{
				{
					"id":             tmdbShow.ID,
					"name":           tmdbShow.Name,
					"original_name":  tmdbShow.OriginalName,
					"first_air_date": tmdbShow.FirstAirDate,
					"poster_path":    tmdbShow.PosterPath,
					"backdrop_path":  tmdbShow.BackdropPath,
					"overview":       tmdbShow.Overview,
					"vote_average":   tmdbShow.VoteAverage,
					"popularity":     tmdbShow.Popularity,
				},
			},
			"total_pages":   1,
			"total_results": 1,
		}
		c.JSON(http.StatusOK, dto.Success(result))
		return
	}

	// 否则按名称搜索
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	searchResult, err := api.crawler.GetTMDBService().SearchShow(query, page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError("TMDB搜索失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(searchResult))
}

// CrawlShow handles POST /api/v1/crawler/show/:tmdb_id
func (api *CrawlerAPI) CrawlShow(c *gin.Context) {
	tmdbIDStr := c.Param("tmdb_id")
	tmdbID, err := strconv.Atoi(tmdbIDStr)
	if err != nil || tmdbID <= 0 {
		c.JSON(http.StatusBadRequest, dto.BadRequest("invalid tmdb_id"))
		return
	}

	// 添加日志
	fmt.Printf("[CrawlShow] 开始爬取 TMDB ID: %d\n", tmdbID)

	if err := api.crawler.CrawlShow(tmdbID); err != nil {
		fmt.Printf("[CrawlShow] 爬取失败: %v\n", err)
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}

	show, err := api.showRepo.GetByTmdbID(tmdbID)
	if err != nil {
		fmt.Printf("[CrawlShow] 获取show失败: %v\n", err)
		c.JSON(http.StatusInternalServerError, dto.InternalError("failed to load show after crawl"))
		return
	}

	fmt.Printf("[CrawlShow] 爬取成功, show: %+v\n", show)
	c.JSON(http.StatusOK, dto.SuccessWithMessage("Show crawled successfully", show))
}

// RefreshAll handles POST /api/v1/crawler/refresh-all
func (api *CrawlerAPI) RefreshAll(c *gin.Context) {
	if api.taskManager == nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError("task manager not available"))
		return
	}

	task, err := api.taskManager.StartRefreshAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}

	c.JSON(http.StatusAccepted, dto.SuccessWithMessage("Refresh started in background", task))
}

// GetTodayUpdates handles GET /api/v1/crawler/today-updates
func (api *CrawlerAPI) GetTodayUpdates(c *gin.Context) {
	// Get episodes with upload status
	episodes, err := api.episodeRepo.GetTodayUpdatesWithUploadStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(episodes))
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

	successCount := int64(0)
	failedCount := int64(0)
	if api.logRepo != nil {
		successCount, err = api.logRepo.CountByStatus("success")
		if err != nil {
			successCount = 0
		}

		failedCount, err = api.logRepo.CountByStatus("failed")
		if err != nil {
			failedCount = 0
		}
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
		ShowName   string `json:"show_name"`
		PosterPath string `json:"poster_path"`
		ShowStatus string `json:"show_status"`
	}

	result := make([]EpisodeWithShow, 0)
	for _, episode := range episodes {
		item := EpisodeWithShow{
			Episode:    episode,
			ShowName:   episode.Show.Name,
			PosterPath: episode.Show.PosterPath,
			ShowStatus: episode.Show.Status,
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

	if api.taskManager == nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError("task manager not available"))
		return
	}

	task, err := api.taskManager.StartCrawlByStatus(req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}

	c.JSON(http.StatusAccepted, dto.SuccessWithMessage(
		fmt.Sprintf("Started crawling with status: %s", req.Status),
		task,
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
	if api.logRepo == nil {
		c.JSON(http.StatusOK, dto.Success(map[string]interface{}{
			"status":  "unknown",
			"message": "Log repository unavailable",
		}))
		return
	}

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

// GetTask handles GET /api/v1/crawler/tasks/:id
func (api *CrawlerAPI) GetTask(c *gin.Context) {
	if api.taskManager == nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError("task manager not available"))
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.BadRequest("Invalid task ID"))
		return
	}

	task, err := api.taskManager.GetTask(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.NotFound("Task not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(task))
}
