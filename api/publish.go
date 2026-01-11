package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xc9973/go-tmdb-crawler/dto"
	"github.com/xc9973/go-tmdb-crawler/services"
)

// PublishAPI handles publishing endpoints
type PublishAPI struct {
	publisher *services.PublisherService
	markdown  *services.MarkdownService
}

// NewPublishAPI creates a new publish API instance
func NewPublishAPI(
	publisher *services.PublisherService,
	markdown *services.MarkdownService,
) *PublishAPI {
	return &PublishAPI{
		publisher: publisher,
		markdown:  markdown,
	}
}

// PublishTodayUpdates handles POST /api/v1/publish/today
func (api *PublishAPI) PublishTodayUpdates(c *gin.Context) {
	result, err := api.publisher.PublishTodayUpdates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}

	if !result.Success {
		c.JSON(http.StatusBadRequest, dto.BadRequest(result.Error.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessWithMessage("Published successfully", result))
}

// PublishDateRange handles POST /api/v1/publish/range
func (api *PublishAPI) PublishDateRange(c *gin.Context) {
	var req struct {
		StartDate string `json:"start_date" binding:"required"`
		EndDate   string `json:"end_date" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.BadRequest(err.Error()))
		return
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.BadRequest("Invalid start_date format. Use YYYY-MM-DD"))
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.BadRequest("Invalid end_date format. Use YYYY-MM-DD"))
		return
	}

	result, err := api.publisher.PublishDateRange(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}

	if !result.Success {
		c.JSON(http.StatusBadRequest, dto.BadRequest(result.Error.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessWithMessage("Published successfully", result))
}

// PublishShow handles POST /api/v1/publish/show/:id
func (api *PublishAPI) PublishShow(c *gin.Context) {
	idStr := c.Param("id")
	showID, err := parseID(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.BadRequest("Invalid show ID"))
		return
	}

	result, err := api.publisher.PublishShow(showID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}

	if !result.Success {
		c.JSON(http.StatusBadRequest, dto.BadRequest(result.Error.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessWithMessage("Published successfully", result))
}

// PublishWeekly handles POST /api/v1/publish/weekly
func (api *PublishAPI) PublishWeekly(c *gin.Context) {
	result, err := api.publisher.PublishWeeklyUpdates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}

	if !result.Success {
		c.JSON(http.StatusBadRequest, dto.BadRequest(result.Error.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessWithMessage("Published successfully", result))
}

// PublishMonthly handles POST /api/v1/publish/monthly
func (api *PublishAPI) PublishMonthly(c *gin.Context) {
	result, err := api.publisher.PublishMonthlyUpdates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}

	if !result.Success {
		c.JSON(http.StatusBadRequest, dto.BadRequest(result.Error.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessWithMessage("Published successfully", result))
}

// GenerateMarkdownToday handles GET /api/v1/publish/markdown/today
func (api *PublishAPI) GenerateMarkdownToday(c *gin.Context) {
	markdown, err := api.markdown.GenerateTodayUpdates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}

	c.Header("Content-Type", "text/markdown; charset=utf-8")
	c.String(http.StatusOK, markdown)
}

// GenerateMarkdownShow handles GET /api/v1/publish/markdown/show/:id
func (api *PublishAPI) GenerateMarkdownShow(c *gin.Context) {
	idStr := c.Param("id")
	showID, err := parseID(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.BadRequest("Invalid show ID"))
		return
	}

	markdown, err := api.markdown.GenerateShowDetail(showID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}

	c.Header("Content-Type", "text/markdown; charset=utf-8")
	c.String(http.StatusOK, markdown)
}

// GenerateMarkdownRange handles GET /api/v1/publish/markdown/range
func (api *PublishAPI) GenerateMarkdownRange(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, dto.BadRequest("start_date and end_date are required"))
		return
	}

	// Parse dates
	_, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.BadRequest("Invalid start_date format. Use YYYY-MM-DD"))
		return
	}

	_, err = time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.BadRequest("Invalid end_date format. Use YYYY-MM-DD"))
		return
	}

	// This would need episodeRepo access
	// For now, return a placeholder
	c.JSON(http.StatusOK, dto.SuccessWithMessage(
		"Markdown generation for date range - to be implemented",
		map[string]string{
			"start_date": startDateStr,
			"end_date":   endDateStr,
		},
	))
}

// GenerateMarkdownWeekly handles GET /api/v1/publish/markdown/weekly
func (api *PublishAPI) GenerateMarkdownWeekly(c *gin.Context) {
	markdown, err := api.markdown.GenerateWeeklyUpdates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}

	c.Header("Content-Type", "text/markdown; charset=utf-8")
	c.String(http.StatusOK, markdown)
}

// parseID parses a string ID to uint
func parseID(idStr string) (uint, error) {
	var id uint
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		return 0, err
	}
	return id, nil
}
