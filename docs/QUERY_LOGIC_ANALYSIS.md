# TMDB Crawler 查询逻辑梳理

## 目录
1. [架构概览](#架构概览)
2. [数据模型](#数据模型)
3. [查询流程](#查询流程)
4. [刷新功能完整流程](#刷新功能完整流程)
5. [性能分析](#性能分析)
6. [已知问题与优化建议](#已知问题与优化建议)

---

## 架构概览

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   前端 API   │────▶│   后端 API   │────▶│  Repository │────▶│   Database  │
│  (浏览器)    │     │  (Gin 路由)  │     │  (GORM ORM)  │     │  (SQLite)   │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
        │                     │                     │
        │                     ▼                     │
        │              ┌─────────────┐             │
        │              │   Services  │             │
        │              │ (业务逻辑层)  │             │
        │              └─────────────┘             │
        │                     │                     │
        ▼                     ▼                     ▼
   api.js          api/show.go        repositories/*.go
```

---

## 数据模型

### Shows 表 (剧集)
```sql
CREATE TABLE shows (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tmdb_id INTEGER UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    original_name VARCHAR(255),
    status VARCHAR(50),  -- 'Ended'/'Returning Series'/'Canceled'
    first_air_date DATE,
    last_crawled_at TIMESTAMP,
    -- 其他字段...
);
```

### Episodes 表 (集数)
```sql
CREATE TABLE episodes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    show_id INTEGER NOT NULL,
    season_number INTEGER NOT NULL,
    episode_number INTEGER NOT NULL,
    name VARCHAR(255),
    air_date DATE,
    -- 唯一约束
    UNIQUE(show_id, season_number, episode_number)
);
```

**关键**: Episodes 通过 `(show_id, season_number, episode_number)` 唯一标识，而非主键 ID。

---

## 查询流程

### 1. 剧集列表查询

**前端调用**:
```javascript
api.getShows(page, pageSize, search, status)
```

**后端处理** (`api/show.go:ListShows`):
```
GET /api/v1/shows?page=1&page_size=25&status=&search=
```

**Repository 查询** (`repositories/show.go:listWithFilters`):
```go
// 1. 构建基础查询
query := r.db.Model(&models.Show{})

// 2. 应用状态过滤
if status != "" {
    query = query.Where("status = ?", status)
}

// 3. 应用搜索过滤
if search != "" {
    // SQLite: LOWER() + LIKE
    query = query.Where("LOWER(name) LIKE ? OR LOWER(original_name) LIKE ?",
                       "%"+search+"%", "%"+search+"%")
}

// 4. 先统计总数
query.Count(&total)

// 5. 分页查询
query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&shows)
```

**SQL 执行**:
```sql
-- Count 查询
SELECT COUNT(*) FROM shows WHERE status = ? AND (LOWER(name) LIKE ? ...);

-- 数据查询
SELECT * FROM shows WHERE status = ? AND (LOWER(name) LIKE ? ...)
ORDER BY created_at DESC LIMIT 25 OFFSET 0;
```

### 2. 刷新单个剧集

**前端调用**:
```javascript
api.refreshShow(showId)  // showId 是数据库 ID
```

**后端处理** (`api/show.go:RefreshShow`):
```go
// 1. 根据 show_id 获取剧集 (获得 tmdb_id)
show, err := api.showRepo.GetByID(uint(id))

// 2. 调用 Crawler 刷新数据
api.crawler.CrawlShow(show.TmdbID)

// 3. 再次查询获取更新后的数据
updatedShow, err := api.showRepo.GetByID(uint(id))
```

---

## 刷新功能完整流程

### Step 1: 获取 TMDB 数据

**services/crawler.go:CrawlShow**

```go
// 1.1 从 TMDB API 获取剧集详情
tmdbShow, err := s.tmdb.GetShowDetails(tmdbID)
// 返回: {ID, Name, Status, Seasons: [...]}

// 1.2 检查本地数据库是否存在该 show
show, err := s.showRepo.GetByTmdbID(tmdbID)
// SELECT * FROM shows WHERE tmdb_id = ? LIMIT 1
```

### Step 2: 获取所有季/集数据

```go
// 2.1 遍历所有季
for _, season := range tmdbShow.Seasons {
    // 2.2 从 TMDB API 获取该季的所有集
    tmdbSeason, err := s.tmdb.GetSeasonEpisodes(tmdbID, season.SeasonNumber)
    // 返回: {Episodes: [{SeasonNumber, EpisodeNumber, Name, ...}]}

    // 2.3 构建 Episode 对象 (注意: 此时 ID=0)
    episodes = append(episodes, &models.Episode{
        SeasonNumber:  tmdbEpisode.SeasonNumber,
        EpisodeNumber: tmdbEpisode.EpisodeNumber,
        Name:          tmdbEpisode.Name,
        // ID 未设置，默认为 0
    })
}
```

### Step 3: 写入剧集信息

```go
// 3.1 判断是新剧还是更新
if isNewShow {
    s.showRepo.Create(show)     // INSERT INTO shows
} else {
    s.showRepo.Update(show)     // UPDATE shows SET ...
}
```

### Step 4: 写入集数信息 (关键步骤)

**repositories/episode.go:CreateBatch**

```go
func (r *episodeRepository) CreateBatch(episodes []*models.Episode) error {
    return r.db.Transaction(func(tx *gorm.DB) error {
        // 4.1 查询该 show 的所有现有 episodes
        var existing []*models.Episode
        tx.Where("show_id = ?", showID).Find(&existing)
        // SELECT * FROM episodes WHERE show_id = ?

        // 4.2 构建哈希表用于快速查找
        existingMap := make(map[string]*models.Episode)
        for _, ep := range existing {
            key := fmt.Sprintf("%d:%d", ep.SeasonNumber, ep.EpisodeNumber)
            // 例如: "1:5" = Season 1, Episode 5
            existingMap[key] = ep
        }

        // 4.3 逐个处理: 更新已存在的，创建不存在的
        for _, ep := range episodes {
            key := fmt.Sprintf("%d:%d", ep.SeasonNumber, ep.EpisodeNumber)
            if existingEp, found := existingMap[key] {
                // === 已存在: 执行 UPDATE ===
                ep.ID = existingEp.ID           // 设置 ID 使 GORM 执行 UPDATE
                ep.CreatedAt = existingEp.CreatedAt
                tx.Save(ep)
                // UPDATE episodes SET name=?, overview=?, ...
                // WHERE id = ? (使用主键)
            } else {
                // === 不存在: 执行 INSERT ===
                tx.Create(ep)
                // INSERT INTO episodes (show_id, season_number, ...) VALUES (...)
            }
        }
        return nil
    })
}
```

**SQL 执行示例**:

假设 show_id=1，刷新时有 3 个新集数:

```sql
-- 4.1 查询现有集数
SELECT * FROM episodes WHERE show_id = 1;
-- 返回: [{id:10, show_id:1, season_number:1, episode_number:1, ...},
--        {id:11, show_id:1, season_number:1, episode_number:2, ...}]

-- 4.3 处理新集数
-- S01E01 已存在 (id=10) -> UPDATE episodes SET ... WHERE id=10
-- S01E02 已存在 (id=11) -> UPDATE episodes SET ... WHERE id=11
-- S01E03 不存在        -> INSERT INTO episodes (show_id, season_number, episode_number, ...) VALUES (1, 1, 3, ...)
```

### Step 5: 更新元数据

```go
// 5.1 更新 LastCrawledAt 等字段
show.LastCrawledAt = time.Now()
show.LastSeasonNumber = lastSeason.SeasonNumber
show.LastEpisodeCount = lastSeason.EpisodeCount
s.showRepo.Update(show)
// UPDATE shows SET last_crawled_at=?, last_season_number=?, ...
// WHERE id = ?
```

---

## 性能分析

### 当前实现的查询次数 (刷新一个剧集)

| 步骤 | 查询 | 说明 |
|------|------|------|
| 1.1 | 1 次 | SELECT shows WHERE tmdb_id=? |
| 4.1 | 1 次 | SELECT episodes WHERE show_id=? |
| 4.3 | N 次 | UPDATE/INSERT 逐个集数 (N=集数) |
| 5.1 | 1 次 | UPDATE shows SET metadata WHERE id=? |
| **总计** | **N+3 次** | 对于 20 集的剧集 ≈ 23 次查询 |

### 瓶颈点

1. **逐个 Save (4.3)**: N 次数据库写入
   - 对于 100 集的剧集 = 100 次写入
   - SQLite 事务内批量写入会更快

2. **先查询后写入 (4.1 + 4.3)**: 两阶段操作
   - 必须先查询才知道哪些已存在
   - 能否使用 UPSERT 简化?

---

## 已知问题与优化建议

### 问题 1: 逐个 Save 性能差

**现状**:
```go
for _, ep := range episodes {
    tx.Save(ep)  // N 次写入
}
```

**优化方案**: 使用原生 SQL UPSERT (需要数据库支持)
```sql
INSERT INTO episodes (show_id, season_number, episode_number, name, ...)
VALUES (1, 1, 1, 'Episode 1', ...),
       (1, 1, 2, 'Episode 2', ...)
ON CONFLICT(show_id, season_number, episode_number)
DO UPDATE SET name=excluded.name, overview=excluded.overview, ...;
```

**限制**: SQLite 需要特定语法，PostgreSQL 支持更好。

### 问题 2: 先查询后写入的两阶段操作

**现状**:
```
Query existing → Build map → Query + Write per episode
```

**优化可能性**:
- 如果能保证 episodes 完全替换 (不保留历史)，可以先 DELETE 后批量 INSERT
- 但当前需要保留 `created_at`，所以需要区分更新/创建

### 问题 3: GORM Save 的行为

**陷阱**: `Save()` 只根据主键 ID 判断 UPDATE/INSERT
```go
// ❌ 错误理解
ep.ID = 0
Save(ep)  // 会 INSERT，但可能违反唯一约束!

// ✅ 正确做法
if existing {
    ep.ID = existing.ID  // 必须设置 ID
}
Save(ep)  // 才会 UPDATE
```

---

## 总结

### 刷新功能的完整数据流

```
用户点击刷新按钮
    │
    ▼
POST /api/v1/shows/123/refresh
    │
    ▼
showRepo.GetByID(123)  →  获取 tmdb_id=94997
    │
    ▼
crawler.CrawlShow(94997)
    │
    ├─▶ tmdb.GetShowDetails(94997)  ───────▶ TMDB API
    │                                        │
    ├─▶ showRepo.GetByTmdbID(94997)     │
    │                                        │
    ├─▶ tmdb.GetSeasonEpisodes(94997, 1) ─┼──▶ TMDB API (第1季)
    │   tmdb.GetSeasonEpisodes(94997, 2) ─┼──▶ TMDB API (第2季)
    │   ...                                  │
    │                                        │
    ├─▶ episodeRepo.CreateBatch([...])     │
    │   │                                    │
    │   ├─▶ SELECT * FROM episodes          │
    │   │    WHERE show_id=123               │
    │   │                                    │
    │   ├─▶ For each episode:               │
    │   │    if exists: UPDATE               │
    │   │    else: INSERT                   │
    │   │                                    │
    │   └─▶ showRepo.Update(show)           │
    │                                        │
    └─▶ showRepo.GetByID(123)  ◀────────────┘
         返回更新后的数据给前端
```

### 关键要点

1. **唯一约束**: Episodes 通过 `(show_id, season_number, episode_number)` 唯一标识
2. **两阶段操作**: 先查询现有 → 再逐个更新/创建
3. **事务保护**: CreateBatch 在事务中执行，保证原子性
4. **GORM Save**: 必须设置正确的 ID 才能执行 UPDATE

### 待优化项

- [ ] 考虑使用原生 SQL UPSERT 提升性能
- [ ] 大量集数时分批处理
- [ ] 添加查询性能监控
