package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/xc9973/go-tmdb-crawler/utils"
)

// CacheService defines the interface for caching operations
type CacheService interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
	GetOrSet(ctx context.Context, key string, dest interface{}, ttl time.Duration, fn func() (interface{}, error)) error
	InvalidatePattern(ctx context.Context, pattern string) error
}

// memoryCacheService implements in-memory caching using go-cache
type memoryCacheService struct {
	cache      *cache.Cache
	logger     *utils.Logger
	stats      *cacheStats
	defaultTTL time.Duration
}

type cacheStats struct {
	hits    int64
	misses  int64
	sets    int64
	deletes int64
}

// NewMemoryCacheService creates a new in-memory cache service
func NewMemoryCacheService(defaultTTL time.Duration, logger *utils.Logger) CacheService {
	// Create cache with default cleanup interval of 10 minutes
	c := cache.New(5*time.Minute, 10*time.Minute)

	return &memoryCacheService{
		cache:      c,
		logger:     logger,
		stats:      &cacheStats{},
		defaultTTL: defaultTTL,
	}
}

// Get retrieves a value from cache
func (s *memoryCacheService) Get(ctx context.Context, key string, dest interface{}) error {
	value, found := s.cache.Get(key)
	if !found {
		s.stats.misses++
		s.logger.Debug("Cache miss", "key", key)
		return fmt.Errorf("cache miss")
	}

	s.stats.hits++
	s.logger.Debug("Cache hit", "key", key)

	// Unmarshal based on destination type
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cached value: %w", err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("failed to unmarshal cached value: %w", err)
	}

	return nil
}

// Set stores a value in cache
func (s *memoryCacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if ttl == 0 {
		ttl = s.defaultTTL
	}

	s.cache.Set(key, value, ttl)
	s.stats.sets++
	s.logger.Debug("Cache set", "key", key, "ttl", ttl)
	return nil
}

// Delete removes a value from cache
func (s *memoryCacheService) Delete(ctx context.Context, key string) error {
	s.cache.Delete(key)
	s.stats.deletes++
	s.logger.Debug("Cache delete", "key", key)
	return nil
}

// Clear clears all cache entries
func (s *memoryCacheService) Clear(ctx context.Context) error {
	s.cache.Flush()
	s.logger.Info("Cache cleared")
	return nil
}

// GetOrSet retrieves a value from cache or sets it using the provided function
func (s *memoryCacheService) GetOrSet(ctx context.Context, key string, dest interface{}, ttl time.Duration, fn func() (interface{}, error)) error {
	// Try to get from cache first
	if err := s.Get(ctx, key, dest); err == nil {
		return nil
	}

	// Cache miss, call the function to get the value
	value, err := fn()
	if err != nil {
		return fmt.Errorf("failed to execute cache function: %w", err)
	}

	// Store in cache
	if err := s.Set(ctx, key, value, ttl); err != nil {
		s.logger.Warn("Failed to cache value", "key", key, "error", err)
	}

	// Set the destination value
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return nil
}

// InvalidatePattern removes all cache entries matching a pattern
// Note: This is a simple implementation that iterates through all keys
// For production, consider using Redis with pattern matching support
func (s *memoryCacheService) InvalidatePattern(ctx context.Context, pattern string) error {
	items := s.cache.Items()
	count := 0

	for key := range items {
		// Simple pattern matching (supports * wildcard)
		if matchPattern(key, pattern) {
			s.cache.Delete(key)
			count++
		}
	}

	s.logger.Info("Cache pattern invalidated", "pattern", pattern, "count", count)
	return nil
}

// GetStats returns cache statistics
func (s *memoryCacheService) GetStats() map[string]interface{} {
	total := s.stats.hits + s.stats.misses
	hitRate := float64(0)
	if total > 0 {
		hitRate = float64(s.stats.hits) / float64(total) * 100
	}

	return map[string]interface{}{
		"hits":     s.stats.hits,
		"misses":   s.stats.misses,
		"sets":     s.stats.sets,
		"deletes":  s.stats.deletes,
		"hit_rate": fmt.Sprintf("%.2f%%", hitRate),
		"items":    s.cache.ItemCount(),
	}
}

// matchPattern checks if a key matches a pattern with * wildcard
func matchPattern(key, pattern string) bool {
	// Simple wildcard matching
	if pattern == "*" {
		return true
	}

	// Convert pattern to regex-like matching
	// For simplicity, just check prefix/suffix
	if len(pattern) == 0 {
		return key == ""
	}

	// Check for exact match
	if pattern == key {
		return true
	}

	// Check for prefix match (pattern*)
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			return true
		}
	}

	// Check for suffix match (*pattern)
	if len(pattern) > 0 && pattern[0] == '*' {
		suffix := pattern[1:]
		if len(key) >= len(suffix) && key[len(key)-len(suffix):] == suffix {
			return true
		}
	}

	return false
}

// CacheKeyBuilder helps build consistent cache keys
type CacheKeyBuilder struct {
	prefix string
}

// NewCacheKeyBuilder creates a new cache key builder
func NewCacheKeyBuilder(prefix string) *CacheKeyBuilder {
	return &CacheKeyBuilder{prefix: prefix}
}

// Build creates a cache key from parts
func (b *CacheKeyBuilder) Build(parts ...string) string {
	key := b.prefix
	for _, part := range parts {
		key += ":" + part
	}
	return key
}

// Common cache key builders
var (
	// ShowCacheKeyBuilder builds cache keys for shows
	ShowCacheKeyBuilder = NewCacheKeyBuilder("show")

	// EpisodeCacheKeyBuilder builds cache keys for episodes
	EpisodeCacheKeyBuilder = NewCacheKeyBuilder("episode")

	// TodayCacheKeyBuilder builds cache keys for today's updates
	TodayCacheKeyBuilder = NewCacheKeyBuilder("today")

	// SearchCacheKeyBuilder builds cache keys for search results
	SearchCacheKeyBuilder = NewCacheKeyBuilder("search")
)

// Common TTL values
const (
	// CacheTTLShort is for frequently changing data (5 minutes)
	CacheTTLShort = 5 * time.Minute

	// CacheTTLMedium is for moderately changing data (15 minutes)
	CacheTTLMedium = 15 * time.Minute

	// CacheTTLLong is for rarely changing data (1 hour)
	CacheTTLLong = 1 * time.Hour

	// CacheTTLVeryLong is for static data (24 hours)
	CacheTTLVeryLong = 24 * time.Hour
)
