# go-tmdb-crawler 项目逻辑架构

## 一、整体分层架构

```
┌─────────────────────────────────────────────────────────────┐
│                        表现层 (Presentation)                   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │   Web UI     │  │  CLI 命令    │  │  API 路由     │       │
│  │  (HTML/JS)   │  │  (Cobra)     │  │  (Gin)       │       │
│  └──────────────┘  └──────────────┘  └──────────────┘       │
└─────────────────────────────────────────────────────────────┘
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                        业务逻辑层 (Service)                    │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐         │
│  │ Crawler  │ │ TMDB     │ │Publisher │ │ Telegraph │         │
│  │ Service  │ │ Service  │ │ Service  │ │ Service  │         │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘         │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐         │
│  │Markdown  │ │  Cache   │ │  Auth    │ │ Backup   │         │
│  │ Service  │ │ Service  │ │ Service  │ │ Service  │         │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘         │
└─────────────────────────────────────────────────────────────┘
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                        数据访问层 (Repository)                 │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐         │
│  │  Show    │ │ Episode  │ │ CrawlLog │ │  Task    │         │
│  │   Repo   │ │   Repo   │ │   Repo   │ │   Repo   │         │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘         │
└─────────────────────────────────────────────────────────────┘
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                        数据层 (Data)                          │
│              ┌─────────────┐    ┌─────────────┐              │
│              │   SQLite    │    │  Telegraph  │              │
│              │  (GORM)     │    │     API     │              │
│              └─────────────┘    └─────────────┘              │
└─────────────────────────────────────────────────────────────┘
                              ▲
┌─────────────────────────────────────────────────────────────┐
│                        外部服务层                             │
│              ┌─────────────┐    ┌─────────────┐              │
│              │   TMDB API  │    │ Telegraph   │              │
│              │             │    │     API     │              │
│              └─────────────┘    └─────────────┘              │
└─────────────────────────────────────────────────────────────┘
```

---

## 二、核心数据流

### 1. 爬取流程

```
用户请求 (API/CLI)
      │
      ▼
┌─────────────────┐
│  ShowAPI.Create │ ──► CrawlerService.CrawlShow(tmdbID)
│  ShowAPI.Refresh│
└─────────────────┘
      │
      ▼
┌─────────────────────────────────────────────────────────────┐
│                    Phase 1: 数据获取                         │
│  1. TMDBService.GetShowDetails(tmdbID)                      │
│  2. TMDBService.GetSeasonEpisodes(tmdbID, season)           │
│     (所有季节数据在内存中准备好)                              │
└─────────────────────────────────────────────────────────────┘
      │
      ▼
┌─────────────────────────────────────────────────────────────┐
│                    Phase 2: 数据准备                         │
│  1. 检查 Show 是否存在                                      │
│  2. 准备 Show 对象 (新增或更新)                             │
│  3. 准备所有 Episode 对象                                   │
└─────────────────────────────────────────────────────────────┘
      │
      ▼
┌─────────────────────────────────────────────────────────────┐
│                    Phase 3: 数据写入                         │
│  DB.Transaction:                                            │
│    1. ShowRepo.Create/Update(show)                          │
│    2. EpisodeRepo.CreateBatch(episodes)                     │
│       └── DELETE FROM episodes WHERE show_id = ?            │
│       └── INSERT INTO episodes VALUES (...)                 │
└─────────────────────────────────────────────────────────────┘
      │
      ▼
   清理缓存 + 记录日志
```

### 2. 发布流程

```
用户请求发布今日更新
      │
      ▼
┌─────────────────────────────────────────────────────────────┐
│  PublisherService.PublishTodayUpdates()                     │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ 1. EpisodeRepo.GetTodayUpdates()                     │   │
│  │    └── 查询今天播出的 episodes                         │   │
│  │ 2. MarkdownService.GenerateTodayUpdates()            │   │
│  │    └── 生成 Markdown 内容                             │   │
│  │ 3. TelegraphService.CreatePage()                     │   │
│  │    └── 发布到 Telegraph                               │   │
│  │ 4. TelegraphPostRepo.Create()                        │   │
│  │    └── 保存发布记录                                   │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

---

## 三、模块职责

| 层级 | 模块 | 职责 |
|------|------|------|
| **API** | `ShowAPI` | 剧集 CRUD、刷新操作 |
| | `CrawlerAPI` | 爬取状态、日志查询 |
| | `PublishAPI` | 发布到 Telegraph |
| | `SchedulerAPI` | 定时任务控制 |
| **Service** | `CrawlerService` | 核心：爬取编排、数据一致性 |
| | `TMDBService` | TMDB API 调用、缓存 |
| | `PublisherService` | 发布流程编排 |
| | `TelegraphService` | Telegraph API 封装 |
| | `MarkdownService` | 内容生成 |
| | `CacheService` | 内存缓存 (15分钟 TTL) |
| **Repository** | `ShowRepository` | 剧集数据访问 |
| | `EpisodeRepository` | 剧集数据访问 (批量事务) |
| | `CrawlLogRepository` | 爬取日志 |
| | `TelegraphPostRepository` | 发布记录 |
| **Model** | `Show` | 剧集实体 (含刷新判断逻辑) |
| | `Episode` | 剧集实体 (含验证钩子) |

---

## 四、关键设计决策

### 1. 数据一致性保证

**Episode 存储策略** (`repositories/episode.go:59-82`):
```go
// 事务保证原子性
db.Transaction(func(tx *gorm.DB) error {
    // 先删除旧数据 → 避免残留
    tx.Where("show_id = ?", showID).Delete(&Episode{})
    // 批量插入新数据
    tx.CreateInBatches(episodes, 100)
    return nil
})
```

**爬取策略** (`services/crawler.go:53-209`):
- 先获取所有 TMDB 数据到内存
- 再一次性写入数据库
- 避免"部分获取 + 部分写入"导致的数据不一致

### 2. 缓存策略

| 缓存类型 | 位置 | TTL | 失效策略 |
|---------|------|-----|---------|
| API 响应缓存 | `CacheService` | 15分钟 | 数据变更时主动失效 |
| TMDB API 缓存 | `TMDBService` | 5分钟 | 定期清理过期 |

### 3. 刷新判断逻辑

`Show.ShouldRefresh()` (`models/show.go:109-119`):
- **Returning Series**: 超过 24 小时刷新
- **Ended Shows**: 超过 7 天刷新
- **Never Crawled**: 立即刷新

---

## 五、目录结构

```
go-tmdb-crawler/
├── api/           # HTTP 处理器、路由设置
├── cmd/           # CLI 命令
├── config/        # 配置加载
├── dto/           # 数据传输对象
├── middleware/    # 认证中间件
├── models/        # 数据模型 (GORM)
├── repositories/  # 数据访问层
├── services/      # 业务逻辑层
│   └── backup/   # 备份服务
├── utils/         # 工具函数
├── web/           # 静态前端文件
└── migrations/    # 数据库迁移脚本
```

---

## 六、入口流程

```
main.go (如存在)
    │
    ▼
cmd/server.go → api.SetupRouter()
                    │
                    ▼
         ┌──────────────────────┐
         │ 1. 加载配置          │
         │ 2. 打开数据库        │
         │ 3. 初始化 Repos      │
         │ 4. 初始化 Services   │
         │ 5. 初始化 APIs       │
         │ 6. 注册路由          │
         │ 7. 启动调度器(可选)  │
         └──────────────────────┘
                    │
                    ▼
              gin.Run(addr)
```

---

## 七、时序图：刷新剧集

```
User          ShowAPI      CrawlerService   TMDBService    DB/Repo
 │              │               │               │            │
 │──Refresh───▶ │               │               │            │
 │              │               │               │            │
 │              │──CrawlShow──▶ │               │            │
 │              │               │               │            │
 │              │               │──GetShow──────────────────▶│
 │              │               │               │            │
 │              │               │──GetSeasons────────────────▶│
 │              │               │               │            │
 │              │               │──Create/Update────────────▶│
 │              │               │               │            │
 │              │               │──CreateBatch──────────────▶│
 │              │               │   (事务)     │            │
 │              │               │               │            │
 │              │◀─────────────│               │            │
 │◀─────────────│               │               │            │
```

---

## 八、API 端点总览

### 公开路由 (无需认证)

| 方法 | 路径 | 功能 |
|------|------|------|
| GET | `/api/v1/shows` | 剧集列表 |
| GET | `/api/v1/shows/:id` | 剧集详情 |
| GET | `/api/v1/shows/:id/episodes` | 剧集列表 |
| GET | `/api/v1/calendar/today` | 今日更新 |
| GET | `/api/v1/crawler/updates` | 日期范围更新 |
| GET | `/api/v1/crawler/status` | 爬取状态 |
| GET | `/api/v1/crawler/search/tmdb` | 搜索 TMDB |

### 管理员路由 (需认证)

| 方法 | 路径 | 功能 |
|------|------|------|
| POST | `/api/v1/shows` | 创建剧集 |
| PUT | `/api/v1/shows/:id` | 更新剧集 |
| DELETE | `/api/v1/shows/:id` | 删除剧集 |
| POST | `/api/v1/shows/:id/refresh` | 刷新剧集 |
| POST | `/api/v1/crawler/show/:tmdb_id` | 爬取剧集 |
| POST | `/api/v1/crawler/refresh-all` | 刷新所有 |
| POST | `/api/v1/publish/today` | 发布今日更新 |
| POST | `/api/v1/scheduler/crawl-now` | 立即爬取 |
| GET | `/api/v1/backup/export` | 导出备份 |

---

## 九、总结

这是一个典型的**三层架构 + 仓储模式**的 Go Web 应用：

1. **表现层**: Gin + 静态 Web UI
2. **业务层**: Service 编排核心逻辑
3. **数据层**: Repository + GORM 抽象数据访问
4. **外部集成**: TMDB API + Telegraph API

### 核心亮点

- 事务保证数据一致性
- 先获取后写入避免部分失败
- 多级缓存提升性能
- 清晰的职责分离
