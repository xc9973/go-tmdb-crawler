# 今日更新上传追踪功能实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**目标:** 在"今日更新"页面为每个剧集添加"已上传"标记功能，用户可标记已上传到 NAS 的剧集，避免重复检查。

**架构:** 新建 `uploaded_episodes` 表持久化标记状态，Repository 层封装数据操作，API 层提供标记/取消接口，前端在今日更新页面添加对号按钮，点击后调用 API 更新状态。

**技术栈:** Go 1.24, GORM, Gin, SQLite, Bootstrap 5, Vanilla JavaScript

**关键说明:**
- 数据库使用 SQLite (INTEGER PRIMARY KEY AUTOINCREMENT)
- API 响应格式: `code: 0` 表示成功，`code: 400/404/500` 表示错误
- 管理员接口需要使用 AdminAuthMiddleware
- 迁移文件 005 已存在，直接使用

---

## Task 1: 数据库迁移 - 创建 uploaded_episodes 表 (已完成)

**文件:**
- 已存在: `migrations/005_add_uploaded_episodes.sql`

**Step 1: 验证迁移文件已存在**

Run: `ls -la migrations/005_add_uploaded_episodes.sql`
Expected: 文件存在

**Step 2: 应用迁移到开发数据库**

Run: `sqlite3 tmdb.db < migrations/005_add_uploaded_episodes.sql`
Expected: 无错误，表创建成功

**Step 3: 验证表结构**

Run: `sqlite3 tmdb.db ".schema uploaded_episodes"`
Expected: 显示完整的表结构和索引

---

## Task 2: Model 层 - 定义 UploadedEpisode 模型

**文件:**
- 创建: `models/uploaded_episode.go`
- 测试: `models/uploaded_episode_test.go`

**Step 1: 编写失败的测试**

```go
// models/uploaded_episode_test.go
package models

import (
    "testing"
    "time"
)

func TestUploadedEpisode_Validate(t *testing.T) {
    tests := []struct {
        name    string
        episode *UploadedEpisode
        wantErr bool
    }{
        {
            name: "valid episode",
            episode: &UploadedEpisode{
                EpisodeID: 1,
                Uploaded:  true,
            },
            wantErr: false,
        },
        {
            name: "missing episode_id",
            episode: &UploadedEpisode{
                Uploaded: true,
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if err := tt.episode.Validate(); (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

**Step 2: 运行测试验证失败**

Run: `go test ./models -run TestUploadedEpisode_Validate -v`
Expected: FAIL with "undefined: UploadedEpisode"

**Step 3: 编写最小实现**

```go
// models/uploaded_episode.go
package models

import (
    "errors"
    "time"
)

// UploadedEpisode tracks which episodes have been uploaded to NAS
type UploadedEpisode struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    EpisodeID uint      `gorm:"not null;uniqueIndex" json:"episode_id"`
    Uploaded  bool      `gorm:"not null;default:true" json:"uploaded"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (UploadedEpisode) TableName() string {
    return "uploaded_episodes"
}

// BeforeCreate hook sets timestamps
func (ue *UploadedEpisode) BeforeCreate(tx *gorm.DB) error {
    now := time.Now()
    ue.CreatedAt = now
    ue.UpdatedAt = now
    return ue.Validate()
}

// BeforeUpdate hook sets updated_at
func (ue *UploadedEpisode) BeforeUpdate(tx *gorm.DB) error {
    ue.UpdatedAt = time.Now()
    return ue.Validate()
}

// Validate validates the uploaded episode data
func (ue *UploadedEpisode) Validate() error {
    if ue.EpisodeID == 0 {
        return errors.New("episode ID cannot be empty")
    }
    return nil
}
```

**Step 4: 运行测试验证通过**

Run: `go test ./models -run TestUploadedEpisode_Validate -v`
Expected: PASS

**Step 5: 提交**

```bash
git add models/uploaded_episode.go models/uploaded_episode_test.go
git commit -m "feat: add UploadedEpisode model"
```

---

## Task 3: Repository 层 - 数据访问接口

**文件:**
- 创建: `repositories/uploaded_episode.go`
- 测试: `repositories/uploaded_episode_test.go`

**Step 1: 编写失败的测试**

```go
// repositories/uploaded_episode_test.go
package repositories

import (
    "testing"
    "time"

    "github.com/xc9973/go-tmdb-crawler/models"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func setupUploadedEpisodeDB(t *testing.T) *gorm.DB {
    dbName := "file:UploadedEpisodeTest_?mode=memory&cache=shared"
    db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
    if err != nil {
        t.Fatalf("Failed to open test database: %v", err)
    }

    err = db.AutoMigrate(&models.UploadedEpisode{}, &models.Episode{}, &models.Show{})
    if err != nil {
        t.Fatalf("Failed to migrate test database: %v", err)
    }

    return db
}

func TestUploadedEpisodeRepository_MarkUploaded(t *testing.T) {
    db := setupUploadedEpisodeDB(t)
    repo := NewUploadedEpisodeRepository(db)

    // Create test episode
    show := &models.Show{TmdbID: 1, Name: "Test Show"}
    db.Create(show)
    episode := &models.Episode{ShowID: show.ID, SeasonNumber: 1, EpisodeNumber: 1}
    db.Create(episode)

    // Test marking as uploaded
    err := repo.MarkUploaded(episode.ID)
    if err != nil {
        t.Fatalf("MarkUploaded() error = %v", err)
    }

    // Verify
    var ue models.UploadedEpisode
    err = db.Where("episode_id = ?", episode.ID).First(&ue).Error
    if err != nil {
        t.Fatalf("Failed to find uploaded episode: %v", err)
    }

    if !ue.Uploaded {
        t.Error("Expected uploaded to be true")
    }
}

func TestUploadedEpisodeRepository_IsUploaded(t *testing.T) {
    db := setupUploadedEpisodeDB(t)
    repo := NewUploadedEpisodeRepository(db)

    // Create test episode
    show := &models.Show{TmdbID: 1, Name: "Test Show"}
    db.Create(show)
    episode := &models.Episode{ShowID: show.ID, SeasonNumber: 1, EpisodeNumber: 1}
    db.Create(episode)

    // Not uploaded initially
    uploaded, err := repo.IsUploaded(episode.ID)
    if err != nil {
        t.Fatalf("IsUploaded() error = %v", err)
    }
    if uploaded {
        t.Error("Expected uploaded to be false initially")
    }

    // Mark as uploaded
    repo.MarkUploaded(episode.ID)

    // Now should be uploaded
    uploaded, err = repo.IsUploaded(episode.ID)
    if err != nil {
        t.Fatalf("IsUploaded() error = %v", err)
    }
    if !uploaded {
        t.Error("Expected uploaded to be true after marking")
    }
}
```

**Step 2: 运行测试验证失败**

Run: `go test ./repositories -run TestUploadedEpisode -v`
Expected: FAIL with "undefined: NewUploadedEpisodeRepository"

**Step 3: 编写最小实现**

```go
// repositories/uploaded_episode.go
package repositories

import (
    "github.com/xc9973/go-tmdb-crawler/models"
    "gorm.io/gorm"
    "gorm.io/gorm/clause"
)

// UploadedEpisodeRepository defines the interface for uploaded episode operations
type UploadedEpisodeRepository interface {
    MarkUploaded(episodeID uint) error
    UnmarkUploaded(episodeID uint) error
    IsUploaded(episodeID uint) (bool, error)
    GetByEpisodeID(episodeID uint) (*models.UploadedEpisode, error)
}

type uploadedEpisodeRepository struct {
    db *gorm.DB
}

// NewUploadedEpisodeRepository creates a new uploaded episode repository
func NewUploadedEpisodeRepository(db *gorm.DB) UploadedEpisodeRepository {
    return &uploadedEpisodeRepository{db: db}
}

// MarkUploaded marks an episode as uploaded (idempotent)
func (r *uploadedEpisodeRepository) MarkUploaded(episodeID uint) error {
    return r.db.Clauses(clause.OnConflict{
        Columns:   []clause.Column{{Name: "episode_id"}},
        DoUpdates: clause.AssignmentColumns([]string{"uploaded", "updated_at"}),
    }).Create(&models.UploadedEpisode{
        EpisodeID: episodeID,
        Uploaded:  true,
    }).Error
}

// UnmarkUploaded removes the uploaded mark for an episode
func (r *uploadedEpisodeRepository) UnmarkUploaded(episodeID uint) error {
    return r.db.Where("episode_id = ?", episodeID).Delete(&models.UploadedEpisode{}).Error
}

// IsUploaded checks if an episode is marked as uploaded
func (r *uploadedEpisodeRepository) IsUploaded(episodeID uint) (bool, error) {
    var count int64
    err := r.db.Model(&models.UploadedEpisode{}).
        Where("episode_id = ? AND uploaded = ?", episodeID, true).
        Count(&count).Error
    return count > 0, err
}

// GetByEpisodeID retrieves the upload record for an episode
func (r *uploadedEpisodeRepository) GetByEpisodeID(episodeID uint) (*models.UploadedEpisode, error) {
    var ue models.UploadedEpisode
    err := r.db.Where("episode_id = ?", episodeID).First(&ue).Error
    if err != nil {
        return nil, err
    }
    return &ue, nil
}
```

**Step 4: 运行测试验证通过**

Run: `go test ./repositories -run TestUploadedEpisode -v`
Expected: PASS

**Step 5: 提交**

```bash
git add repositories/uploaded_episode.go repositories/uploaded_episode_test.go
git commit -m "feat: add UploadedEpisodeRepository"
```

---

## Task 4: API 层 - 今日更新接口返回上传状态

**文件:**
- 修改: `api/crawler.go`
- 修改: `repositories/episode.go`

**Step 1: 在 repositories/episode.go 中添加带上传状态的查询方法**

```go
// repositories/episode.go

// GetTodayUpdatesWithUploadStatus retrieves episodes airing today with upload status
// 返回结构与前端 today.js 期望的格式匹配
func (r *episodeRepository) GetTodayUpdatesWithUploadStatus() ([]map[string]interface{}, error) {
    start, end := r.timezoneHelper.TodayRange()

    type Result struct {
        ID            uint
        SeasonNumber  int
        EpisodeNumber int
        Name          string
        AirDate       *time.Time
        StillPath     string
        VoteAverage   float32
        ShowID        uint
        ShowName      string
        PosterPath    string
        ShowStatus    string
        Uploaded      bool
    }

    var results []Result
    err := r.db.Raw(`
        SELECT
            e.id,
            e.season_number,
            e.episode_number,
            e.name,
            e.air_date,
            e.still_path,
            e.vote_average,
            e.show_id,
            s.name as show_name,
            s.poster_path,
            s.status as show_status,
            COALESCE(ue.uploaded, 0) as uploaded
        FROM episodes e
        INNER JOIN shows s ON e.show_id = s.id
        LEFT JOIN uploaded_episodes ue ON e.id = ue.episode_id
        WHERE e.air_date >= ? AND e.air_date < ?
        ORDER BY e.air_date ASC
    `, start, end).Scan(&results).Error

    if err != nil {
        return nil, err
    }

    // 转换为 map 格式以匹配前端期望
    episodes := make([]map[string]interface{}, len(results))
    for i, r := range results {
        episodes[i] = map[string]interface{}{
            "id":             r.ID,
            "season_number":  r.SeasonNumber,
            "episode_number": r.EpisodeNumber,
            "name":           r.Name,
            "air_date":       r.AirDate,
            "still_path":     r.StillPath,
            "vote_average":   r.VoteAverage,
            "show_id":        r.ShowID,
            "show_name":      r.ShowName,
            "poster_path":    r.PosterPath,
            "show_status":    r.ShowStatus,
            "uploaded":       r.Uploaded,
        }
    }

    return episodes, nil
}
```

**Step 2: 修改 api/crawler.go 的 GetTodayUpdates 方法**

```go
// api/crawler.go

// GetTodayUpdates handles GET /api/v1/calendar/today
func (api *CrawlerAPI) GetTodayUpdates(c *gin.Context) {
    // Cast to access new method
    episodeRepoImpl, ok := api.episodeRepo.(*repositories.episodeRepository)
    if !ok {
        c.JSON(http.StatusInternalServerError, dto.InternalError("Repository type mismatch"))
        return
    }

    // Get episodes with upload status
    episodes, err := episodeRepoImpl.GetTodayUpdatesWithUploadStatus()
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.InternalError(err.Error()))
        return
    }

    c.JSON(http.StatusOK, dto.Success(episodes))
}
```

**Step 3: 验证编译**

Run: `go build ./...`
Expected: 编译成功

**Step 4: 提交**

```bash
git add repositories/episode.go api/crawler.go
git commit -m "feat: include upload status in today updates API"
```

---

## Task 5: API 层 - 标记/取消标记接口

**文件:**
- 创建: `api/uploaded_episode.go`
- 修改: `api/setup.go`

**Step 1: 创建 UploadedEpisode API**

```go
// api/uploaded_episode.go
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
```

**Step 2: 在 setup.go 中注册路由和初始化**

修改 `api/setup.go`:

1. 在依赖注入部分 (line 73 附近) 添加:
```go
uploadedEpisodeRepo := repositories.NewUploadedEpisodeRepository(db)
```

2. 在 API 初始化部分 (line 116 附近) 添加:
```go
uploadedEpisodeAPI := NewUploadedEpisodeAPI(episodeRepo, uploadedEpisodeRepo)
```

3. 在管理员路由组 (line 154-199) 添加:
```go
// Episode upload tracking (write operations - requires admin auth)
admin.POST("/episodes/:id/uploaded", uploadedEpisodeAPI.MarkUploaded)
admin.DELETE("/episodes/:id/uploaded", uploadedEpisodeAPI.UnmarkUploaded)
```

**Step 3: 验证编译**

Run: `go build ./...`
Expected: 编译成功，无错误

**Step 4: 提交**

```bash
git add api/uploaded_episode.go api/setup.go
git commit -m "feat: add episode upload tracking API endpoints"
```

---

## Task 6: 前端 API 客户端 - 添加标记接口

**文件:**
- 修改: `web/js/common.js`

**Step 1: 在 APIClient 类中添加方法**

在 `web/js/common.js` 的 `APIClient` 类中 (line 544 之后) 添加:

```javascript
// ========== Episode Upload Tracking ==========

/**
 * 标记剧集已上传
 * POST /api/v1/episodes/:id/uploaded
 */
async markEpisodeUploaded(episodeId) {
    return this.post(`/episodes/${episodeId}/uploaded`, {});
}

/**
 * 取消标记剧集已上传
 * DELETE /api/v1/episodes/:id/uploaded
 */
async unmarkEpisodeUploaded(episodeId) {
    return this.delete(`/episodes/${episodeId}/uploaded`);
}
```

**Step 2: 验证无语法错误**

Run: 在浏览器控制台检查或启动应用测试

**Step 3: 提交**

```bash
git add web/js/common.js
git commit -m "feat: add upload tracking API methods to client"
```

---

## Task 7: 前端 UI - 今日更新页面添加对号按钮

**文件:**
- 修改: `web/js/today.js`
- 修改: `web/css/custom.css`

**Step 1: 在 custom.css 中添加对号按钮样式**

在 `web/css/custom.css` 文件末尾 (line 174 之后) 添加:

```css
/* 上传标记按钮 */
.upload-check-btn {
    width: 28px;
    height: 28px;
    border-radius: 50%;
    border: 2px solid #dee2e6;
    background: rgba(255, 255, 255, 0.9);
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    transition: all 0.2s ease;
    margin-left: auto;
    flex-shrink: 0;
}

.upload-check-btn:hover {
    border-color: #198754;
    transform: scale(1.1);
}

.upload-check-btn.uploaded {
    background: #198754;
    border-color: #198754;
    color: white;
}

.upload-check-btn.uploaded::before {
    content: "\F26B"; /* Bootstrap Icons check-lg */
    font-family: "bootstrap-icons";
    font-size: 14px;
}

.upload-check-btn:not(.uploaded)::before {
    content: "\F26E"; /* Bootstrap Icons circle */
    font-family: "bootstrap-icons";
    font-size: 14px;
    color: #6c757d;
}

.upload-check-btn.loading {
    opacity: 0.6;
    pointer-events: none;
}
```

**Step 2: 修改 today.js 的 renderShows 方法**

修改 `web/js/today.js` 中的 `renderShows` 方法 (line 106-177):

在 episodesHTML 生成部分 (line 130 左右)，修改为:

```javascript
show.episodes.slice(0, 5).forEach(ep => {
    const episodeCode = `S${ep.season_number}E${ep.episode_number}`;
    const isUploaded = ep.uploaded || false;
    const checkBtnClass = isUploaded ? 'uploaded' : '';
    const btnTitle = isUploaded ? '已上传 - 点击取消' : '标记已上传';

    episodesHTML += `
        <li class="d-flex align-items-center gap-2">
            <button class="upload-check-btn ${checkBtnClass}"
                    data-episode-id="${ep.id}"
                    onclick="todayPage.toggleUploaded(${ep.id}, event)"
                    title="${btnTitle}">
            </button>
            <span class="flex-grow-1">
                <i class="bi bi-play-circle"></i> ${episodeCode} - ${this.escapeHtml(ep.name)}
            </span>
        </li>`;
});
```

**Step 3: 在 TodayPage 类中添加 toggleUploaded 方法**

在 `web/js/today.js` 的 `TodayPage` 类中添加新方法 (line 370 左右，`refreshShow` 方法之后):

```javascript
/**
 * 切换剧集的上传状态
 * @param {number} episodeId - 剧集ID
 * @param {Event} event - 点击事件
 */
async toggleUploaded(episodeId, event) {
    event.stopPropagation(); // 防止触发父元素点击

    const btn = event.target.closest('.upload-check-btn');
    if (!btn) return;

    const isCurrentlyUploaded = btn.classList.contains('uploaded');

    // 乐观更新 UI
    btn.classList.add('loading');
    if (isCurrentlyUploaded) {
        btn.classList.remove('uploaded');
    } else {
        btn.classList.add('uploaded');
    }

    try {
        // 调用 API
        if (isCurrentlyUploaded) {
            await api.unmarkEpisodeUploaded(episodeId);
            btn.title = '标记已上传';
        } else {
            await api.markEpisodeUploaded(episodeId);
            btn.title = '已上传 - 点击取消';
        }
        this.showSuccess(isCurrentlyUploaded ? '已取消标记' : '已标记为上传');
    } catch (error) {
        // 失败回滚
        if (isCurrentlyUploaded) {
            btn.classList.add('uploaded');
            btn.title = '已上传 - 点击取消';
        } else {
            btn.classList.remove('uploaded');
            btn.title = '标记已上传';
        }
        this.showError('操作失败: ' + error.message);
    } finally {
        btn.classList.remove('loading');
    }
}
```

**Step 4: 在浏览器中验证功能**

Run: 打开浏览器访问 /today.html
Expected: 每个剧集旁显示圆形对号按钮，点击可切换状态

**Step 5: 提交**

```bash
git add web/js/today.js web/css/custom.css
git commit -m "feat: add upload check button in today updates page"
```

---

## Task 8: 数据库迁移执行与验证

**Step 1: 应用迁移到开发数据库**

Run: `sqlite3 tmdb.db < migrations/005_add_uploaded_episodes.sql`
Expected: 无错误

**Step 2: 验证表创建**

Run: `sqlite3 tmdb.db "SELECT name FROM sqlite_master WHERE type='table' AND name='uploaded_episodes';"`
Expected: 返回 `uploaded_episodes`

**Step 3: 验证表结构**

Run: `sqlite3 tmdb.db ".schema uploaded_episodes"`
Expected: 显示完整的表结构和索引

---

## Task 9: 端到端测试与验证

**Step 1: 启动应用**

Run: `go run main.go`
Expected: 服务启动，监听默认端口 (8080)

**Step 2: 手动测试流程**

1. 访问 http://localhost:8080/today.html
2. 登录管理员账户
3. 确认每个剧集旁显示圆形按钮
4. 点击按钮，观察状态变化 (未上传 → 已上传)
5. 刷新页面，确认状态持久化
6. 再次点击，取消标记 (已上传 → 未上传)
7. 打开开发者工具，查看 API 请求和响应

**Step 3: API 测试 (使用管理员 Cookie)**

```bash
# 1. 先登录获取 cookie
curl -c cookies.txt -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"api_key":"your-admin-api-key"}'

# 2. 标记已上传
curl -b cookies.txt -X POST http://localhost:8080/api/v1/episodes/1/uploaded

# 3. 查询今日更新 (应包含 uploaded 字段)
curl -b cookies.txt http://localhost:8080/api/v1/calendar/today

# 4. 取消标记
curl -b cookies.txt -X DELETE http://localhost:8080/api/v1/episodes/1/uploaded
```

Expected JSON 响应格式:
```json
// 标记成功响应
{
  "code": 0,
  "message": "success",
  "data": {
    "uploaded": true,
    "episode_id": 1
  }
}

// 今日更新响应 (包含 uploaded 字段)
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "season_number": 1,
      "episode_number": 1,
      "name": "Episode Name",
      "show_id": 100,
      "show_name": "Show Name",
      "uploaded": true
    }
  ]
}
```

**Step 4: 运行所有测试**

Run: `go test ./... -v`
Expected: 所有测试通过

**Step 5: 清理测试 cookies**

Run: `rm cookies.txt`

---

## 验证清单

### 后端
- [ ] 数据库表 `uploaded_episodes` 创建成功
- [ ] Model 层 UploadedEpisode 定义正确
- [ ] Repository 层 UploadedEpisodeRepository 测试通过
- [ ] API 接口返回正确的 JSON 格式 (code: 0 表示成功)
- [ ] 管理员中间件正确应用

### 前端
- [ ] 前端按钮显示正确 (未上传显示圆圈，已上传显示对号)
- [ ] 点击后状态立即更新 (乐观 UI)
- [ ] API 失败时正确回滚状态
- [ ] 刷新页面状态保持持久化
- [ ] 按钮悬停显示正确提示文本

### 集成
- [ ] 所有单元测试通过
- [ ] 手动端到端测试通过
- [ ] API 认证正常工作

---

## 故障排查

### 问题: API 返回 401 Unauthorized
**原因**: 未登录或 Cookie 过期
**解决**: 确保已登录管理员账户

### 问题: 前端按钮点击无反应
**原因**: JavaScript 错误或 API 路由未注册
**解决**: 检查浏览器控制台错误日志，确认 setup.go 中路由已注册

### 问题: uploaded 字段始终为 false
**原因**: 数据库未迁移或 LEFT JOIN 未正确执行
**解决**: 确认已执行迁移，检查 repositories/episode.go 中的 SQL

### 问题: 刷新后状态丢失
**原因**: 数据库事务未提交或 SQLite 文件锁定
**解决**: 检查数据库文件权限，确认应用有写权限

---

## 参考文档

- 设计文档: `docs/plans/2026-01-14-uploaded-tracking-design.md`
- 现有代码模式: `models/episode.go`, `repositories/episode.go`
- API 响应格式: `dto/response.go`
- 认证中间件: `middleware/auth.go`
- 前端 API 客户端: `web/js/common.js`

---

**End of Implementation Plan**
