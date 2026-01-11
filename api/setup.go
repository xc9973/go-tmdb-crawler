package api

import (
	"github.com/gin-gonic/gin"
	"github.com/xc9973/go-tmdb-crawler/config"
	"github.com/xc9973/go-tmdb-crawler/middleware"
)

// SetupRouter creates and configures the Gin router
func SetupRouter(cfg *config.Config) *gin.Engine {
	router := gin.Default()

	// CORS middleware
	router.Use(corsMiddleware(cfg))

	// Serve static files
	router.Static("/css", cfg.Paths.Web+"/css")
	router.Static("/js", cfg.Paths.Web+"/js")
	router.StaticFile("/", cfg.Paths.Web+"/index.html")
	router.StaticFile("/today.html", cfg.Paths.Web+"/today.html")
	router.StaticFile("/logs.html", cfg.Paths.Web+"/logs.html")
	router.StaticFile("/show_detail.html", cfg.Paths.Web+"/show_detail.html")

	// API routes
	api := router.Group("/api/v1")
	{
		// 公开路由 - 无需认证
		// Shows (只读)
		api.GET("/shows", getShows)
		api.GET("/shows/:id", getShow)

		// Calendar (只读)
		api.GET("/calendar/today", getTodayUpdates)
		api.GET("/calendar", getCalendar)

		// Crawler (只读状态)
		api.GET("/crawler/status", getCrawlerStatus)
	}

	// 管理员路由 - 需要认证
	admin := router.Group("/api/v1")
	admin.Use(middleware.AdminAuthMiddleware())
	{
		// Shows (写操作)
		admin.POST("/shows", createShow)
		admin.PUT("/shows/:id", updateShow)
		admin.DELETE("/shows/:id", deleteShow)
		admin.POST("/shows/:id/refresh", refreshShow)

		// Crawler (写操作和日志)
		admin.POST("/crawler/show/:tmdb_id", crawlShow)
		admin.GET("/crawler/logs", getCrawlLogs)
	}

	return router
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

// Placeholder handlers - to be implemented
func getShows(c *gin.Context) {
	c.JSON(200, gin.H{"code": 0, "message": "success", "data": gin.H{"items": []interface{}{}, "total": 0}})
}

func getShow(c *gin.Context) {
	c.JSON(200, gin.H{"code": 0, "message": "success", "data": gin.H{}})
}

func createShow(c *gin.Context) {
	c.JSON(200, gin.H{"code": 0, "message": "success"})
}

func updateShow(c *gin.Context) {
	c.JSON(200, gin.H{"code": 0, "message": "success"})
}

func deleteShow(c *gin.Context) {
	c.JSON(200, gin.H{"code": 0, "message": "success"})
}

func refreshShow(c *gin.Context) {
	c.JSON(200, gin.H{"code": 0, "message": "success"})
}

func crawlShow(c *gin.Context) {
	c.JSON(200, gin.H{"code": 0, "message": "success"})
}

func getCrawlLogs(c *gin.Context) {
	c.JSON(200, gin.H{"code": 0, "message": "success", "data": gin.H{"items": []interface{}{}, "total": 0}})
}

func getCrawlerStatus(c *gin.Context) {
	c.JSON(200, gin.H{"code": 0, "message": "success", "data": gin.H{"status": "idle"}})
}

func getTodayUpdates(c *gin.Context) {
	c.JSON(200, gin.H{"code": 0, "message": "success", "data": gin.H{"updates": []interface{}{}}})
}

func getCalendar(c *gin.Context) {
	c.JSON(200, gin.H{"code": 0, "message": "success", "data": gin.H{"days": []interface{}{}}})
}
