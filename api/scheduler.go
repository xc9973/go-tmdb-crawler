package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xc9973/go-tmdb-crawler/dto"
	"github.com/xc9973/go-tmdb-crawler/services"
)

// SchedulerAPI handles scheduler control endpoints
type SchedulerAPI struct {
	scheduler *services.Scheduler
}

// NewSchedulerAPI creates a new scheduler API instance
func NewSchedulerAPI(scheduler *services.Scheduler) *SchedulerAPI {
	return &SchedulerAPI{
		scheduler: scheduler,
	}
}

// GetStatus handles GET /api/v1/scheduler/status
func (api *SchedulerAPI) GetStatus(c *gin.Context) {
	status := api.scheduler.GetStatus()
	c.JSON(http.StatusOK, dto.Success(status))
}

// GetNextRunTimes handles GET /api/v1/scheduler/next-runs
func (api *SchedulerAPI) GetNextRunTimes(c *gin.Context) {
	nextRuns := api.scheduler.GetNextRunTimes()
	c.JSON(http.StatusOK, dto.Success(nextRuns))
}

// StartScheduler handles POST /api/v1/scheduler/start
func (api *SchedulerAPI) StartScheduler(c *gin.Context) {
	if err := api.scheduler.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.SuccessWithMessage("Scheduler started successfully", nil))
}

// StopScheduler handles POST /api/v1/scheduler/stop
func (api *SchedulerAPI) StopScheduler(c *gin.Context) {
	api.scheduler.Stop()
	c.JSON(http.StatusOK, dto.SuccessWithMessage("Scheduler stopped successfully", nil))
}

// RunCrawlNow handles POST /api/v1/scheduler/crawl-now
func (api *SchedulerAPI) RunCrawlNow(c *gin.Context) {
	if err := api.scheduler.RunCrawlNow(); err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.SuccessWithMessage("Crawl job triggered successfully", nil))
}

// RunPublishNow handles POST /api/v1/scheduler/publish-now
func (api *SchedulerAPI) RunPublishNow(c *gin.Context) {
	result, err := api.scheduler.RunPublishNow()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.SuccessWithMessage("Publish job triggered successfully", result))
}

// RunManualCrawl handles POST /api/v1/scheduler/crawl/:id
func (api *SchedulerAPI) RunManualCrawl(c *gin.Context) {
	var req struct {
		ID int `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.BadRequest(err.Error()))
		return
	}

	if err := api.scheduler.RunManualCrawl(req.ID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.SuccessWithMessage("Manual crawl completed successfully", nil))
}

// RunManualPublish handles POST /api/v1/scheduler/publish/:id
func (api *SchedulerAPI) RunManualPublish(c *gin.Context) {
	var req struct {
		ID uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.BadRequest(err.Error()))
		return
	}

	result, err := api.scheduler.RunManualPublish(req.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.SuccessWithMessage("Manual publish completed successfully", result))
}

// GetTimeouts handles GET /api/v1/scheduler/timeouts
func (api *SchedulerAPI) GetTimeouts(c *gin.Context) {
	timeouts := api.scheduler.GetTimeouts()
	c.JSON(http.StatusOK, dto.Success(timeouts))
}

// SetTimeouts handles PUT /api/v1/scheduler/timeouts
func (api *SchedulerAPI) SetTimeouts(c *gin.Context) {
	var req struct {
		CrawlTimeout   int `json:"crawl_timeout" binding:"required,min=1"`
		PublishTimeout int `json:"publish_timeout" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.BadRequest(err.Error()))
		return
	}

	api.scheduler.SetTimeouts(
		time.Duration(req.CrawlTimeout)*time.Second,
		time.Duration(req.PublishTimeout)*time.Second,
	)
	c.JSON(http.StatusOK, dto.SuccessWithMessage("Timeouts updated successfully", nil))
}
