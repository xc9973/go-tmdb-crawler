# 今日更新上传追踪功能设计

**日期:** 2026-01-14
**状态:** 设计完成，待实施

## 需求概述

在"今日更新"页面为每个剧集添加"已上传"标记功能，用户可将已上传到 NAS 的剧集打勾，避免重复检查。

## 设计要点

### 1. 数据库设计

新建 `uploaded_episodes` 表：

```sql
CREATE TABLE uploaded_episodes (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    episode_id  INTEGER NOT NULL UNIQUE,
    uploaded    BOOLEAN NOT NULL DEFAULT 1,
    created_at  DATETIME,
    updated_at  DATETIME,
    FOREIGN KEY (episode_id) REFERENCES episodes(id) ON DELETE CASCADE
);

CREATE INDEX idx_uploaded_episodes_episode_id ON uploaded_episodes(episode_id);
```

- 使用 `episode_id` 外键关联现有剧集数据
- `uploaded` 字段预留后续可能需要"取消标记"的场景
- 唯一索引确保每个剧集只有一条记录
- 外键级联删除：剧集删除时自动清理标记记录

### 2. API 设计

| 方法 | 路径 | 功能 |
|------|------|------|
| POST | `/api/episodes/:id/uploaded` | 标记剧集已上传 |
| DELETE | `/api/episodes/:id/uploaded` | 取消标记（可选） |
| GET | `/api/today-updates` | 扩展响应，包含 uploaded 状态 |

**请求/响应格式：**

```go
// POST /api/episodes/:id/uploaded 响应
{
  "success": true,
  "uploaded": true
}

// GET /api/today-updates 响应扩展
{
  "episodes": [{
    "id": 123,
    "series_name": "剧名",
    "season_number": 1,
    "episode_number": 1,
    "title": "剧名 S01E01",
    "uploaded": true,           // 新增字段
    "uploaded_at": "2026-01-14T10:30:00Z"
  }]
}
```

**鉴权：** 复用现有的 JWT 中间件，需要登录后操作。

### 3. 前端交互设计

**UI 改动：**

在"今日更新"页面的每个剧集卡片上添加：
- 右上角圆形复选按钮
  - 未选中：⚪ 半透明边框
  - 已选中：✓ 绿色背景 + 白色对勾
- 按钮尺寸：约 32x32px，不遮挡原有内容

**交互流程：**
1. 点击按钮 → POST 请求 → 成功后立即更新 UI 状态
2. 点击已选中按钮 → DELETE 请求 → 取消标记
3. 错误处理：请求失败时 Toast 提示，按钮恢复原状态

**状态管理：**
- 页面加载时从 `/api/today-updates` 获取所有剧集的 `uploaded` 状态
- 本地维护一个 `Set<episodeId>` 缓存已上传 ID
- 点击时乐观更新 UI，失败后回滚

### 4. 实现要点

**后端文件结构：**

```
models/uploaded_episode.go          # 数据模型
repositories/uploaded_episode.go    # 数据访问层
api/uploaded_episode.go             # API 处理器（或扩展现有）
```

**关键代码模式：**

```go
// Repository 创建或更新
func (r *UploadedEpisodeRepository) MarkUploaded(episodeID int) error {
    return r.db.Clauses(clause.OnConflict{
        Columns:   []clause.Column{{Name: "episode_id"}},
        DoUpdates: clause.AssignmentColumns([]string{"uploaded", "updated_at"}),
    }).Create(&UploadedEpisode{
        EpisodeID: episodeID,
        Uploaded:  true,
    }).Error
}

// 今日更新查询扩展
func (r *EpisodeRepository) GetTodayUpdates() ([]EpisodeWithUploadStatus, error) {
    // LEFT JOIN uploaded_episodes 填充 uploaded 字段
}
```

**前端文件结构：**

```
web/static/js/today-updates.js     # 修改：添加上传状态处理
web/static/css/components.css      # 修改：添加对号按钮样式
```

### 5. 边界情况处理

| 场景 | 处理方式 |
|------|----------|
| 剧集删除 | 数据库外键 `ON DELETE CASCADE` 自动清理 |
| 并发点击 | 后端幂等性保证，使用 `FirstOrCreate` 或 `OnConflict` |
| 网络超时 | 前端 3 秒超时 + 失败重试一次 |
| 重复标记 | 后端返回成功，不创建重复记录 |
| 未登录访问 | JWT 中间件返回 401，前端跳转登录 |

### 6. 测试覆盖

**单元测试：**
- `UploadedEpisodeRepository` CRUD 操作
- API 接口（标记、取消标记）
- 并发标记测试

**集成测试：**
- 标记 → 刷新页面 → 状态保持
- 标记后删除剧集 → 记录自动清理

**边界测试：**
- 重复标记同一剧集
- 取消已取消的标记
- 未登录用户标记

## 约束确认

- [x] 数据库持久化存储
- [x] 已完成显示绿色✓图标，保留在列表中
- [x] 永久保留标记，除非手动取消
- [x] 仅在"今日更新"页面实现
- [x] 最小改动，复用现有能力
- [x] 不新建小函数，复用现有模式
