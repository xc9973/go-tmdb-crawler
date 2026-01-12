package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xc9973/go-tmdb-crawler/utils"
)

// MetricsMiddleware tracks request metrics
type MetricsMiddleware struct {
	logger *utils.Logger
	stats  *RequestStats
}

// RequestStats holds request statistics
type RequestStats struct {
	TotalRequests   int64
	SuccessRequests int64
	ErrorRequests   int64
	AverageLatency  time.Duration
	SlowRequests    int64 // Requests taking > 1s
}

// NewMetricsMiddleware creates a new metrics middleware
func NewMetricsMiddleware(logger *utils.Logger) *MetricsMiddleware {
	return &MetricsMiddleware{
		logger: logger,
		stats:  &RequestStats{},
	}
}

// Middleware returns the Gin middleware function
func (m *MetricsMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Process request
		c.Next()

		// Calculate metrics
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// Update stats
		m.stats.TotalRequests++
		if statusCode >= 200 && statusCode < 400 {
			m.stats.SuccessRequests++
		} else {
			m.stats.ErrorRequests++
		}

		// Track slow requests
		if latency > time.Second {
			m.stats.SlowRequests++
			m.logger.Warn("Slow request detected",
				"path", path,
				"method", method,
				"status", statusCode,
				"latency", latency.String(),
				"ip", c.ClientIP(),
			)
		}

		// Update average latency (simple moving average)
		if m.stats.TotalRequests == 1 {
			m.stats.AverageLatency = latency
		} else {
			m.stats.AverageLatency = (m.stats.AverageLatency*time.Duration(m.stats.TotalRequests-1) + latency) /
				time.Duration(m.stats.TotalRequests)
		}

		// Log request details
		m.logger.Info("Request",
			"method", method,
			"path", path,
			"status", statusCode,
			"latency", latency.String(),
			"ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)
	}
}

// GetStats returns current statistics
func (m *MetricsMiddleware) GetStats() *RequestStats {
	return m.stats
}

// ResetStats resets all statistics
func (m *MetricsMiddleware) ResetStats() {
	m.stats = &RequestStats{}
}

// PerformanceMiddleware adds performance-related headers and monitoring
func PerformanceMiddleware(logger *utils.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Calculate processing time
		duration := time.Since(start)

		// Add performance headers
		c.Header("X-Response-Time", duration.String())
		c.Header("X-Process-Time", strconv.FormatInt(duration.Milliseconds(), 10))

		// Log if response time is high
		if duration > 500*time.Millisecond {
			logger.Warn("High response time",
				"path", c.Request.URL.Path,
				"method", c.Request.Method,
				"duration", duration.String(),
			)
		}
	}
}

// ResponseSizeMiddleware tracks response sizes
func ResponseSizeMiddleware(logger *utils.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request
		c.Next()

		// Get response size
		size := c.Writer.Size()

		// Add header
		c.Header("X-Response-Size", strconv.Itoa(size))

		// Log large responses
		if size > 1024*1024 { // > 1MB
			logger.Warn("Large response detected",
				"path", c.Request.URL.Path,
				"size", strconv.Itoa(size),
			)
		}
	}
}

// ConcurrencyLimitMiddleware limits concurrent requests
type ConcurrencyLimitMiddleware struct {
	semaphore chan struct{}
	logger    *utils.Logger
}

// NewConcurrencyLimitMiddleware creates a new concurrency limiter
func NewConcurrencyLimitMiddleware(maxConcurrent int, logger *utils.Logger) *ConcurrencyLimitMiddleware {
	return &ConcurrencyLimitMiddleware{
		semaphore: make(chan struct{}, maxConcurrent),
		logger:    logger,
	}
}

// Middleware returns the Gin middleware function
func (m *ConcurrencyLimitMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		select {
		case m.semaphore <- struct{}{}:
			defer func() { <-m.semaphore }()
			c.Next()
		default:
			m.logger.Warn("Concurrency limit reached",
				"path", c.Request.URL.Path,
				"ip", c.ClientIP(),
			)
			c.JSON(429, gin.H{
				"error": "Too many concurrent requests",
				"code":  "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
		}
	}
}
