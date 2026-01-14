package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xc9973/go-tmdb-crawler/dto"
	"github.com/xc9973/go-tmdb-crawler/models"
	"github.com/xc9973/go-tmdb-crawler/services/backup"
)

const (
	maxBackupFileSize = 50 * 1024 * 1024 // 50MB
)

// BackupAPI handles backup-related API endpoints
type BackupAPI struct {
	backupService backup.Service
}

// NewBackupAPI creates a new backup API instance
func NewBackupAPI(backupService backup.Service) *BackupAPI {
	return &BackupAPI{
		backupService: backupService,
	}
}

// ExportBackup handles GET /api/v1/backup/export
func (api *BackupAPI) ExportBackup(c *gin.Context) {
	// Export data
	backupData, err := api.backupService.Export()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(fmt.Sprintf("导出失败: %s", err.Error())))
		return
	}

	// Marshal to JSON with indentation
	jsonData, err := json.MarshalIndent(backupData, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError(fmt.Sprintf("JSON序列化失败: %s", err.Error())))
		return
	}

	// Generate filename
	filename := fmt.Sprintf("tmdb-backup-%s.json", time.Now().Format("20060102-150405"))

	// Set headers for file download
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Data(http.StatusOK, "application/json", jsonData)
}

// ImportBackup handles POST /api/v1/backup/import
func (api *BackupAPI) ImportBackup(c *gin.Context) {
	// Parse multipart form
	if err := c.Request.ParseMultipartForm(maxBackupFileSize); err != nil {
		c.JSON(http.StatusBadRequest, dto.BadRequest("文件过大或解析失败: "+err.Error()))
		return
	}

	// Get file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.BadRequest("未找到上传文件"))
		return
	}
	defer file.Close()

	// Check file size
	if header.Size > maxBackupFileSize {
		c.JSON(http.StatusBadRequest, dto.BadRequest(fmt.Sprintf("文件过大: %d bytes (最大 %d bytes)", header.Size, maxBackupFileSize)))
		return
	}

	// Read file content
	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError("读取文件失败: "+err.Error()))
		return
	}

	// Validate JSON
	var backupData models.BackupExport
	if err := json.Unmarshal(data, &backupData); err != nil {
		c.JSON(http.StatusBadRequest, dto.BadRequest("JSON格式错误: "+err.Error()))
		return
	}

	// Validate version
	if backupData.Version != models.BackupVersion {
		c.JSON(http.StatusBadRequest, dto.BadRequest(fmt.Sprintf("不支持的备份版本: %s (当前支持: %s)", backupData.Version, models.BackupVersion)))
		return
	}

	// Get import mode
	mode := backup.ImportMode(c.DefaultPostForm("mode", "replace"))
	if mode != backup.ImportModeReplace && mode != backup.ImportModeMerge {
		c.JSON(http.StatusBadRequest, dto.BadRequest("无效的导入模式: "+string(mode)))
		return
	}

	// Import data
	result, err := api.backupService.Import(&backupData, mode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError("导入失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessWithMessage("导入成功", result))
}

// GetBackupStatus handles GET /api/v1/backup/status
func (api *BackupAPI) GetBackupStatus(c *gin.Context) {
	status, err := api.backupService.GetStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalError("获取备份状态失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(status))
}
