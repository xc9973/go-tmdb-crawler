package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xc9973/go-tmdb-crawler/dto"
	"github.com/xc9973/go-tmdb-crawler/repositories"
	"github.com/xc9973/go-tmdb-crawler/services/correction"
)

// CorrectionAPI handles correction-related endpoints
type CorrectionAPI struct {
	correction *correction.Service
	showRepo   repositories.ShowRepository
}

// NewCorrectionAPI creates a new correction API instance
func NewCorrectionAPI(
	correctionService *correction.Service,
	showRepo repositories.ShowRepository,
) *CorrectionAPI {
	return &CorrectionAPI{
		correction: correctionService,
		showRepo:   showRepo,
	}
}

// GetStatus handles GET /api/v1/correction/status
func (api *CorrectionAPI) GetStatus(c *gin.Context) {
	result := api.correction.GetLastDetectionResult()
	if result == nil {
		// No cached result, return empty status
		c.JSON(http.StatusOK, dto.Success(map[string]interface{}{
			"total_shows":     0,
			"stale_count":     0,
			"pending_refresh": 0,
			"duration_ms":     0,
			"stale_shows":     []*correction.StaleShowInfo{},
		}))
		return
	}

	response := map[string]interface{}{
		"total_shows":     result.TotalShowsAnalyzed,
		"stale_count":     result.StaleShowsFound,
		"pending_refresh": result.TasksCreated,
		"duration_ms":     result.Duration.Milliseconds(),
		"stale_shows":     result.StaleShows,
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
	result := api.correction.GetLastDetectionResult()
	if result == nil {
		// No cached result, return empty list
		c.JSON(http.StatusOK, dto.Success([]*correction.StaleShowInfo{}))
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
