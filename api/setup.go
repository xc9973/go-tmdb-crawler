package api

import (
	"log"
	"strings"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/xc9973/go-tmdb-crawler/config"
	"github.com/xc9973/go-tmdb-crawler/middleware"
	"github.com/xc9973/go-tmdb-crawler/models"
	"github.com/xc9973/go-tmdb-crawler/repositories"
	backupservice "github.com/xc9973/go-tmdb-crawler/services/backup"
	"github.com/xc9973/go-tmdb-crawler/services"
	"github.com/xc9973/go-tmdb-crawler/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// mustOpenDB 打开数据库连接，失败时panic
func mustOpenDB(cfg *config.Config) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(cfg.Database.Path), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return db
}

// SetupRouter creates and configures the Gin router
func SetupRouter(cfg *config.Config) *gin.Engine {
	router := gin.Default()

	// CORS middleware
	router.Use(corsMiddleware(cfg))

	// Gzip compression
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	// Static file cache middleware
	router.Use(staticCacheMiddleware())

	// 公开静态文件 (CSS/JS/登录页/欢迎页)
	router.Static("/css", cfg.Paths.Web+"/css")
	router.Static("/js", cfg.Paths.Web+"/js")
	router.StaticFile("/login.html", cfg.Paths.Web+"/login.html")
	router.StaticFile("/welcome.html", cfg.Paths.Web+"/welcome.html")

	// 需要认证的Web页面
	webPages := router.Group("")
	webPages.Use(middleware.WebAuthMiddleware())
	{
		webPages.StaticFile("/", cfg.Paths.Web+"/index.html")
		webPages.StaticFile("/today.html", cfg.Paths.Web+"/today.html")
		webPages.StaticFile("/logs.html", cfg.Paths.Web+"/logs.html")
		webPages.StaticFile("/show_detail.html", cfg.Paths.Web+"/show_detail.html")
		webPages.StaticFile("/backup.html", cfg.Paths.Web+"/backup.html")
	}

	// Initialize timezone helper
	location, err := time.LoadLocation(cfg.Timezone.Default)
	if err != nil {
		log.Fatalf("Failed to load timezone location '%s': %v", cfg.Timezone.Default, err)
	}
	timezoneHelper := utils.NewTimezoneHelper(location)

	// Dependencies
	db := mustOpenDB(cfg)
	showRepo := repositories.NewShowRepository(db)
	episodeRepo := repositories.NewEpisodeRepository(db)
	crawlLogRepo := repositories.NewCrawlLogRepository(db)
	crawlTaskRepo := repositories.NewCrawlTaskRepository(db)
	telegraphPostRepo := repositories.NewTelegraphPostRepository(db)
	uploadedEpisodeRepo := repositories.NewUploadedEpisodeRepository(db)

	// Set timezone helper for episode repository
	episodeRepo.SetTimezoneHelper(timezoneHelper)

	// Auto migrate database tables
	// Note: Episode table structure is managed by SQL migrations (see migrations/001_init_schema.sql)
	// Note: UploadedEpisode table structure is managed by SQL migrations (see migrations/005_add_uploaded_episodes.sql)
	// We only AutoMigrate tables that don't have complex constraints
	if err := db.AutoMigrate(
		&models.Show{},
		// &models.Episode{}, // Skip - managed by SQL migrations
		&models.CrawlLog{},
		&models.CrawlTask{},
		&models.TelegraphPost{},
		&models.Session{},
		// &models.UploadedEpisode{}, // Skip - managed by SQL migrations
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migration completed successfully")

	// 初始化认证服务
	authService := services.NewAuthService(cfg.Auth.SecretKey, db)

	// 设置认证服务到中间件
	middleware.InitAdminAuth(cfg.Auth.SecretKey, cfg.Auth.AllowRemote)
	if adminAuth := middleware.GetAdminAuth(); adminAuth != nil {
		adminAuth.SetAuthService(authService)
	}

	// 初始化认证处理器
	authHandler := NewAuthHandler(authService, middleware.GetAdminAuth())

	telegraph := services.NewTelegraphService(cfg.Telegraph.Token, cfg.Telegraph.AuthorName, cfg.Telegraph.AuthorURL)
	publisher := services.NewPublisherService(telegraph, showRepo, episodeRepo, telegraphPostRepo, timezoneHelper)
	tmdb := services.MustTMDBService(cfg.TMDB.APIKey, cfg.TMDB.BaseURL, cfg.TMDB.Language)
	crawler := services.NewCrawlerService(tmdb, showRepo, episodeRepo, crawlLogRepo, crawlTaskRepo)
	taskManager := services.NewTaskManager(crawlTaskRepo, crawler)

	// Initialize scheduler
	logger := utils.NewLogger(cfg.App.LogLevel, cfg.Paths.Log)
	scheduler := services.NewScheduler(crawler, publisher, logger)

	showAPI := NewShowAPI(showRepo, episodeRepo, crawler)
	crawlerAPI := NewCrawlerAPI(crawler, showRepo, crawlLogRepo, episodeRepo, taskManager)
	markdownService := services.NewMarkdownService(episodeRepo, showRepo)
	markdownService.SetTimezoneHelper(timezoneHelper)
	publishAPI := NewPublishAPI(publisher, markdownService)
	schedulerAPI := NewSchedulerAPI(scheduler)

	// Initialize backup service
	backupService := backupservice.NewService(db, showRepo, episodeRepo, crawlLogRepo, telegraphPostRepo)
	backupAPI := NewBackupAPI(backupService)
	uploadedEpisodeAPI := NewUploadedEpisodeAPI(episodeRepo, uploadedEpisodeRepo)

	// API routes
	api := router.Group("/api/v1")
	{
		// 认证路由 - 公开
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.GET("/session", authHandler.GetSessionInfo)
		}

		// 公开路由 - 无需认证
		// Shows (只读)
		api.GET("/shows", showAPI.ListShows)
		api.GET("/shows/:id", showAPI.GetShow)
		api.GET("/shows/:id/episodes", showAPI.GetShowEpisodes)

		// Calendar (只读)
		api.GET("/calendar/today", crawlerAPI.GetTodayUpdates)
		api.GET("/crawler/updates", crawlerAPI.GetUpdatesByDateRange)

		// Crawler (只读状态)
		api.GET("/crawler/status", crawlerAPI.GetHealthStatus)
		api.GET("/crawler/search/tmdb", crawlerAPI.SearchTMDB)
	}

	// 管理员路由 - 需要认证
	admin := router.Group("/api/v1")
	admin.Use(middleware.AdminAuthMiddleware())
	{
		// Shows (写操作)
		admin.POST("/shows", showAPI.CreateShow)
		admin.PUT("/shows/:id", showAPI.UpdateShow)
		admin.DELETE("/shows/:id", showAPI.DeleteShow)
		admin.POST("/shows/:id/refresh", showAPI.RefreshShow)

		// Crawler (写操作和日志)
		admin.POST("/crawler/show/:tmdb_id", crawlerAPI.CrawlShow)
		admin.POST("/crawler/refresh-all", crawlerAPI.RefreshAll)
		admin.POST("/crawler/crawl-by-status", crawlerAPI.CrawlByStatus)
		admin.GET("/crawler/logs", crawlerAPI.GetCrawlLogs)
		admin.DELETE("/crawler/logs/old", crawlerAPI.DeleteOldLogs)
		admin.GET("/crawler/health", crawlerAPI.GetHealthStatus)
		admin.GET("/crawler/tasks/:id", crawlerAPI.GetTask)

		// Publish
		admin.POST("/publish/today", publishAPI.PublishTodayUpdates)
		admin.POST("/publish/range", publishAPI.PublishDateRange)
		admin.POST("/publish/show/:id", publishAPI.PublishShow)
		admin.POST("/publish/weekly", publishAPI.PublishWeekly)
		admin.POST("/publish/monthly", publishAPI.PublishMonthly)
		admin.GET("/publish/markdown/today", publishAPI.GenerateMarkdownToday)
		admin.GET("/publish/markdown/show/:id", publishAPI.GenerateMarkdownShow)
		admin.GET("/publish/markdown/range", publishAPI.GenerateMarkdownRange)
		admin.GET("/publish/markdown/weekly", publishAPI.GenerateMarkdownWeekly)

		// Scheduler
		admin.GET("/scheduler/status", schedulerAPI.GetStatus)
		admin.GET("/scheduler/next-runs", schedulerAPI.GetNextRunTimes)
		admin.POST("/scheduler/start", schedulerAPI.StartScheduler)
		admin.POST("/scheduler/stop", schedulerAPI.StopScheduler)
		admin.POST("/scheduler/crawl-now", schedulerAPI.RunCrawlNow)
		admin.POST("/scheduler/publish-now", schedulerAPI.RunPublishNow)
		admin.POST("/scheduler/crawl/:id", schedulerAPI.RunManualCrawl)
		admin.POST("/scheduler/publish/:id", schedulerAPI.RunManualPublish)
		admin.GET("/scheduler/timeouts", schedulerAPI.GetTimeouts)
		admin.PUT("/scheduler/timeouts", schedulerAPI.SetTimeouts)

		// Backup
		admin.GET("/backup/export", backupAPI.ExportBackup)
		admin.POST("/backup/import", backupAPI.ImportBackup)
		admin.GET("/backup/status", backupAPI.GetBackupStatus)

		// Episode upload tracking (write operations - requires admin auth)
		admin.POST("/episodes/:id/uploaded", uploadedEpisodeAPI.MarkUploaded)
		admin.DELETE("/episodes/:id/uploaded", uploadedEpisodeAPI.UnmarkUploaded)
	}

	// Start scheduler if enabled
	if cfg.Scheduler.Enabled {
		log.Println("Starting scheduler...")
		if err := scheduler.Start(); err != nil {
			log.Printf("Failed to start scheduler: %v", err)
		} else {
			log.Println("Scheduler started successfully")
		}
	}

	return router
}

// staticCacheMiddleware adds cache headers for static files
func staticCacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		// Remove query string for suffix check
		if idx := strings.Index(path, "?"); idx != -1 {
			path = path[:idx]
		}
		// Add cache headers for CSS and JS files
		if strings.HasSuffix(path, ".css") || strings.HasSuffix(path, ".js") {
			c.Header("Cache-Control", "public, max-age=86400") // 1 day
			c.Header("Vary", "Accept-Encoding")
		}
		c.Next()
	}
}

// corsMiddleware adds CORS headers
func corsMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", cfg.CORS.AllowedOrigins)
		c.Writer.Header().Set("Access-Control-Allow-Methods", cfg.CORS.AllowedMethods)
		c.Writer.Header().Set("Access-Control-Allow-Headers", cfg.CORS.AllowedHeaders)

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
