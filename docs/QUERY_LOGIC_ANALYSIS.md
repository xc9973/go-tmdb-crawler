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

### Step 4: 写入集数信息 (关键步骤 - 已优化)

**repositories/episode.go:CreateBatch** (当前版本)

```go
func (r *episodeRepository) CreateBatch(episodes []*models.Episode) error {
    return r.db.Transaction(func(tx *gorm.DB) error {
        // 4.1 删除该 show 的所有旧 episodes
        tx.Where("show_id = ?", showID).Delete(&models.Episode{})
        // DELETE FROM episodes WHERE show_id = ?

        // 4.2 批量插入所有新 episodes
        tx.CreateInBatches(episodes, 100)
        // INSERT INTO episodes (...) VALUES (...), (...), (...)
        //   GORM 自动批量插入，100 条一批

        return nil
    })
}
```

**优点**:
1. **数据一致性**: TMDB 返回什么，数据库就存什么
2. **性能**: 固定 2 次数据库操作（DELETE + INSERT）
3. **简洁**: 不需要复杂的 upsert 逻辑

**SQL 执行示例**:

假设 show_id=1，TMDB 返回 3 个集数:

```sql
-- 4.1 删除旧集数
DELETE FROM episodes WHERE show_id = 1;
-- 影响: 删除所有旧集数，同时 CASCADE 删除 uploaded_episodes 关联记录

-- 4.2 批量插入新集数
INSERT INTO episodes (show_id, season_number, episode_number, name, ...)
VALUES (1, 1, 1, 'Episode 1', ...),
       (1, 1, 2, 'Episode 2', ...),
       (1, 1, 3, 'Episode 3', ...);
-- 一次性插入，GORM 自动分批（每批100条）
```

**重要说明**:
- `created_at` 会更新为当前时间（新创建的时间戳）
- `uploaded_episodes` 表会通过 `ON DELETE CASCADE` 自动清理关联记录
- 如果 TMDB 删除了某集，数据库中也会被删除（符合预期）

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
| **总计 (旧版)** | **N+3 次** | 对于 20 集的剧集 ≈ 23 次查询 |

### 优化后的性能 (当前版本)

| 步骤 | 查询 | 说明 |
|------|------|------|
| 1.1 | 1 次 | SELECT shows WHERE tmdb_id=? |
| 4.1 | 1 次 | DELETE episodes WHERE show_id=? |
| 4.2 | 1 次 | INSERT INTO episodes (...) VALUES (...) (批量) |
| 5.1 | 1 次 | UPDATE shows SET metadata WHERE id=? |
| **总计 (新版)** | **4 次** | 固定 4 次查询，与集数无关！ |

**性能提升**: 20 集剧集从 23 次查询降低到 4 次查询，提升约 **83%**。

### 瓶颈点

1. **逐个 Save (4.3)**: N 次数据库写入
   - 对于 100 集的剧集 = 100 次写入
   - SQLite 事务内批量写入会更快

2. **先查询后写入 (4.1 + 4.3)**: 两阶段操作
   - 必须先查询才知道哪些已存在
   - 能否使用 UPSERT 简化?

---

## 已知问题与优化建议

### ✅ 已修复的问题

#### 问题 1: 按季调用导致的重复查询 (已修复)

**问题**: 每季调用一次 CreateBatch，每次都查询所有集数
```
第1季: 查询100集 → 处理20集
第2季: 查询100集 → 处理20集
第3季: 查询100集 → 处理20集
第4季: 查询100集 → 处理20集
第5季: 查询100集 → 处理20集
```

**解决方案**: 合并所有季的集数，一次性调用 CreateBatch
```
一次性查询100集 → 一次性处理100集
```

**效果**: 5 次查询 → 1 次查询

#### 问题 2: 旧集数不删除 (已修复)

**问题**: TMDB 删除了某集，但数据库中仍保留
```
数据库: 第1-10集
TMDB:    第1-8集
结果:   第9-10集成为"僵尸"数据
```

**解决方案**: 先删除所有旧集数，再插入新集数
```sql
DELETE FROM episodes WHERE show_id = ?
INSERT INTO episodes (...) VALUES (...)
```

**效果**: 数据库与 TMDB 完全同步

### 潜在的权衡

#### created_at 时间戳

**当前行为**: 刷新剧集时，所有集数的 `created_at` 会更新为当前时间

**影响**:
- ❌ 无法通过 `created_at` 追踪原始添加时间
- ✅ 但数据一致性更重要（与 TMDB 同步）

**如果需要保留原始时间**，可以考虑:
1. 增加字段 `original_created_at`
2. 或使用 upsert 模式（但性能较差）

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
