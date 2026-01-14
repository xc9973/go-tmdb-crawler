package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xc9973/go-tmdb-crawler/dto"
	"github.com/xc9973/go-tmdb-crawler/repositories"
)

// UploadedEpisodeAPI handles episode upload tracking endpoints
type UploadedEpisodeAPI struct {
	episodeRepo repositories.EpisodeRepository
	uploadedRepo repositories.UploadedEpisodeRepository
}

// NewUploadedEpisodeAPI creates a new uploaded episode API instance
func NewUploadedEpisodeAPI(
	episodeRepo repositories.EpisodeRepository,
	uploadedRepo repositories.UploadedEpisodeRepository,
) *UploadedEpisodeAPI {
	return &UploadedEpisodeAPI{
		episodeRepo: episodeRepo,
		uploadedRepo: uploadedRepo,
	}
}

// MarkUploaded handles POST /api/v1/episodes/:id/uploaded
// 需要管理员认证 (AdminAuthMiddleware)
func (api *UploadedEpisodeAPI) MarkUploaded(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.BadRequest("Invalid episode ID"))
		return
	}

	// Verify episode exists
	_, err = api.episodeRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NotFound("Episode not found"))
		return
	}

	// Mark as uploaded
	if err := api.uploadedRepo.MarkUploaded(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(map[string]interface{}{
		"uploaded":    true,
		"episode_id":  uint(id),
	}))
}

// UnmarkUploaded handles DELETE /api/v1/episodes/:id/uploaded
// 需要管理员认证 (AdminAuthMiddleware)
func (api *UploadedEpisodeAPI) UnmarkUploaded(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.BadRequest("Invalid episode ID"))
		return
	}

	if err := api.uploadedRepo.UnmarkUploaded(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(map[string]interface{}{
		"uploaded":    false,
		"episode_id":  uint(id),
	}))
}
