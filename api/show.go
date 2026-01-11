package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/go-tmdb-crawler/dto"
	"github.com/yourusername/go-tmdb-crawler/models"
	"github.com/yourusername/go-tmdb-crawler/repositories"
	"github.com/yourusername/go-tmdb-crawler/services"
)

// ShowAPI handles show-related API endpoints
type ShowAPI struct {
	showRepo repositories.ShowRepository
	crawler  *services.CrawlerService
}

// NewShowAPI creates a new show API instance
func NewShowAPI(showRepo repositories.ShowRepository, crawler *services.CrawlerService) *ShowAPI {
	return &ShowAPI{
		showRepo: showRepo,
		crawler:  crawler,
	}
}

// ListShows handles GET /api/v1/shows
func (api *ShowAPI) ListShows(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")
	search := c.Query("search")

	// Validate page and pageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var shows []*models.Show
	var total int64
	var err error

	// Handle different query types
	if search != "" {
		shows, total, err = api.showRepo.Search(search, page, pageSize)
	} else if status != "" {
		// Filter by status
		shows, total, err = api.showRepo.List(page, pageSize)
		// Manual filtering (in production, add specific repo method)
		filtered := make([]*models.Show, 0)
		for _, show := range shows {
			if show.Status == status {
				filtered = append(filtered, show)
			}
		}
		shows = filtered
		total = int64(len(filtered))
	} else {
		shows, total, err = api.showRepo.List(page, pageSize)
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
		Items:      shows,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, dto.Success(response))
}

// GetShow handles GET /api/v1/shows/:id
func (api *ShowAPI) GetShow(c *gin.Context) {
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

	c.JSON(http.StatusOK, dto.Success(show))
}

// CreateShow handles POST /api/v1/shows
func (api *ShowAPI) CreateShow(c *gin.Context) {
	var req struct {
		TmdbID int `json:"tmdb_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.BadRequest(err.Error()))
		return
	}

	// Crawl the show
	if err := api.crawler.CrawlShow(req.TmdbID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}

	// Get the created show
	show, err := api.showRepo.GetByTmdbID(req.TmdbID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError("Failed to retrieve created show"))
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessWithMessage("Show created successfully", show))
}

// UpdateShow handles PUT /api/v1/shows/:id
func (api *ShowAPI) UpdateShow(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.BadRequest("Invalid show ID"))
		return
	}

	var req models.Show
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.BadRequest(err.Error()))
		return
	}

	// Get existing show
	show, err := api.showRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NotFound("Show not found"))
		return
	}

	// Update fields
	req.ID = uint(id)
	req.TmdbID = show.TmdbID // Keep original TMDB ID
	req.CreatedAt = show.CreatedAt

	if err := api.showRepo.Update(&req); err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessWithMessage("Show updated successfully", &req))
}

// DeleteShow handles DELETE /api/v1/shows/:id
func (api *ShowAPI) DeleteShow(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.BadRequest("Invalid show ID"))
		return
	}

	if err := api.showRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessWithMessage("Show deleted successfully", nil))
}

// RefreshShow handles POST /api/v1/shows/:id/refresh
func (api *ShowAPI) RefreshShow(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.BadRequest("Invalid show ID"))
		return
	}

	// Get show
	show, err := api.showRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NotFound("Show not found"))
		return
	}

	// Refresh show
	if err := api.crawler.CrawlShow(show.TmdbID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}

	// Get updated show
	updatedShow, err := api.showRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError("Failed to retrieve updated show"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessWithMessage("Show refreshed successfully", updatedShow))
}

// BatchCreateShows handles POST /api/v1/shows/batch
func (api *ShowAPI) BatchCreateShows(c *gin.Context) {
	var req struct {
		TmdbIDs []int `json:"tmdb_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.BadRequest(err.Error()))
		return
	}

	// Batch crawl
	results := api.crawler.BatchCrawl(req.TmdbIDs)

	// Count successes
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}

	c.JSON(http.StatusOK, dto.SuccessWithMessage(
		fmt.Sprintf("Batch crawl completed: %d/%d successful", successCount, len(req.TmdbIDs)),
		results,
	))
}

// GetShowEpisodes handles GET /api/v1/shows/:id/episodes
func (api *ShowAPI) GetShowEpisodes(c *gin.Context) {
	_, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.BadRequest("Invalid show ID"))
		return
	}

	// This would need an episode repository
	// For now, return a message
	c.JSON(http.StatusOK, dto.SuccessWithMessage("Episodes feature coming soon", nil))
}
