package services

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/parnurzeal/gorequest"
	"github.com/xc9973/go-tmdb-crawler/dto"
)

// TMDBService handles all TMDB API interactions
type TMDBService struct {
	apiKey     string
	baseURL    string
	lang       string
	client     *gorequest.SuperAgent
	timeout    time.Duration
	maxRetries int
	cache      *TMDBCache
	mu         sync.RWMutex
}

// TMDBCache provides simple in-memory caching for TMDB responses
type TMDBCache struct {
	data map[string]cacheEntry
	mu   sync.RWMutex
	ttl  time.Duration
}

type cacheEntry struct {
	data      []byte
	timestamp time.Time
}

// NewTMDBCache creates a new cache instance
func NewTMDBCache(ttl time.Duration) *TMDBCache {
	cache := &TMDBCache{
		data: make(map[string]cacheEntry),
		ttl:  ttl,
	}
	// Start cleanup goroutine
	go cache.cleanup()
	return cache
}

// Get retrieves data from cache
func (c *TMDBCache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.data[key]
	if !exists {
		return nil, false
	}

	if time.Since(entry.timestamp) > c.ttl {
		return nil, false
	}

	return entry.data, true
}

// Set stores data in cache
func (c *TMDBCache) Set(key string, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = cacheEntry{
		data:      data,
		timestamp: time.Now(),
	}
}

// cleanup removes expired cache entries
func (c *TMDBCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.data {
			if now.Sub(entry.timestamp) > c.ttl {
				delete(c.data, key)
			}
		}
		c.mu.Unlock()
	}
}

// Clear clears all cache entries
func (c *TMDBCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]cacheEntry)
}

// NewTMDBService creates a new TMDB service instance
func NewTMDBService(apiKey, baseURL, lang string) *TMDBService {
	return &TMDBService{
		apiKey:     apiKey,
		baseURL:    baseURL,
		lang:       lang,
		client:     gorequest.New(),
		timeout:    10 * time.Second,
		maxRetries: 3,
		cache:      NewTMDBCache(5 * time.Minute),
	}
}

// NewTMDBServiceWithCache creates a new TMDB service instance with custom cache TTL
func NewTMDBServiceWithCache(apiKey, baseURL, lang string, cacheTTL time.Duration) *TMDBService {
	return &TMDBService{
		apiKey:     apiKey,
		baseURL:    baseURL,
		lang:       lang,
		client:     gorequest.New(),
		timeout:    10 * time.Second,
		maxRetries: 3,
		cache:      NewTMDBCache(cacheTTL),
	}
}

// MustTMDBService returns a TMDB service and panics if API key is missing.
func MustTMDBService(apiKey, baseURL, lang string) *TMDBService {
	if apiKey == "" {
		panic("TMDB_API_KEY is required")
	}
	return NewTMDBService(apiKey, baseURL, lang)
}

// SetTimeout sets the request timeout
func (s *TMDBService) SetTimeout(timeout time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.timeout = timeout
}

// SetMaxRetries sets the maximum number of retries
func (s *TMDBService) SetMaxRetries(maxRetries int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.maxRetries = maxRetries
}

// ClearCache clears the TMDB cache
func (s *TMDBService) ClearCache() {
	s.cache.Clear()
}

// GetCacheStats returns cache statistics
func (s *TMDBService) GetCacheStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"entries": len(s.cache.data),
		"ttl":     s.cache.ttl.String(),
	}
}

// GetShowDetails fetches show details from TMDB
func (s *TMDBService) GetShowDetails(tmdbID int) (*dto.TMDBShowResponse, error) {
	url := fmt.Sprintf("%s/tv/%d", s.baseURL, tmdbID)

	var response dto.TMDBShowResponse
	if err := s.makeRequest(url, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// GetSeasonEpisodes fetches episodes for a specific season
func (s *TMDBService) GetSeasonEpisodes(tmdbID, seasonNumber int) (*dto.TMDBSeasonResponse, error) {
	url := fmt.Sprintf("%s/tv/%d/season/%d", s.baseURL, tmdbID, seasonNumber)

	var response dto.TMDBSeasonResponse
	if err := s.makeRequest(url, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// GetAllSeasons fetches all seasons for a show
func (s *TMDBService) GetAllSeasons(tmdbID int) ([]*dto.TMDBSeasonResponse, error) {
	// First get show details to know how many seasons there are
	show, err := s.GetShowDetails(tmdbID)
	if err != nil {
		return nil, err
	}

	var seasons []*dto.TMDBSeasonResponse
	for _, seasonInfo := range show.Seasons {
		// Skip season 0 (specials)
		if seasonInfo.SeasonNumber == 0 {
			continue
		}

		season, err := s.GetSeasonEpisodes(tmdbID, seasonInfo.SeasonNumber)
		if err != nil {
			return nil, fmt.Errorf("failed to get season %d: %w", seasonInfo.SeasonNumber, err)
		}
		seasons = append(seasons, season)
	}

	return seasons, nil
}

// GetShowWithAllSeasons fetches show details and all seasons
func (s *TMDBService) GetShowWithAllSeasons(tmdbID int) (*dto.TMDBShowResponse, []*dto.TMDBSeasonResponse, error) {
	show, err := s.GetShowDetails(tmdbID)
	if err != nil {
		return nil, nil, err
	}

	seasons, err := s.GetAllSeasons(tmdbID)
	if err != nil {
		return nil, nil, err
	}

	return show, seasons, nil
}

// SearchShow searches for shows by query
func (s *TMDBService) SearchShow(query string, page int) (*dto.TMDBSearchResponse, error) {
	url := fmt.Sprintf("%s/search/tv", s.baseURL)

	var response dto.TMDBSearchResponse
	if err := s.makeRequest(url, &response, map[string]string{
		"query": query,
		"page":  fmt.Sprintf("%d", page),
	}); err != nil {
		return nil, err
	}

	return &response, nil
}

// makeRequest makes an HTTP request to TMDB API with caching and retry logic
func (s *TMDBService) makeRequest(url string, result interface{}, queryParams ...map[string]string) error {
	s.mu.RLock()
	timeout := s.timeout
	maxRetries := s.maxRetries
	s.mu.RUnlock()

	// Generate cache key
	cacheKey := s.generateCacheKey(url, queryParams...)

	// Try to get from cache
	if cachedData, found := s.cache.Get(cacheKey); found {
		if err := json.Unmarshal(cachedData, result); err == nil {
			return nil
		}
	}

	// Build request
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		request := s.client.Get(url).
			Query("api_key=" + s.apiKey).
			Query("language=" + s.lang)

		// Add custom query params if provided
		if len(queryParams) > 0 {
			for key, value := range queryParams[0] {
				request = request.Query(fmt.Sprintf("%s=%s", key, value))
			}
		}

		// Set timeout
		request.Timeout(timeout)

		// Send request
		resp, body, errs := request.EndBytes()
		if len(errs) > 0 {
			lastErr = fmt.Errorf("request failed: %v", errs[0])
			continue
		}

		// Check status code
		if resp.StatusCode != 200 {
			var errResp dto.TMDBErrorResponse
			if err := json.Unmarshal(body, &errResp); err == nil {
				lastErr = fmt.Errorf("TMDB API error: %s", errResp.StatusMessage)
			} else {
				lastErr = fmt.Errorf("HTTP error: %d - %s", resp.StatusCode, string(body))
			}

			// Don't retry on client errors (4xx)
			if resp.StatusCode >= 400 && resp.StatusCode < 500 {
				return lastErr
			}
			continue
		}

		// Parse response
		if err := json.Unmarshal(body, result); err != nil {
			lastErr = fmt.Errorf("failed to parse response: %w", err)
			continue
		}

		// Cache the successful response
		s.cache.Set(cacheKey, body)

		return nil
	}

	return lastErr
}

// generateCacheKey generates a unique cache key for the request
func (s *TMDBService) generateCacheKey(url string, queryParams ...map[string]string) string {
	key := url
	if len(queryParams) > 0 {
		for k, v := range queryParams[0] {
			key += fmt.Sprintf(":%s=%s", k, v)
		}
	}
	return key
}

// GetImageURL returns the full image URL
func (s *TMDBService) GetImageURL(path string, size string) string {
	if path == "" {
		return ""
	}
	return fmt.Sprintf("https://image.tmdb.org/t/p/%s%s", size, path)
}

// GetPosterURL returns the poster image URL
func (s *TMDBService) GetPosterURL(path string) string {
	return s.GetImageURL(path, "w500")
}

// GetBackdropURL returns the backdrop image URL
func (s *TMDBService) GetBackdropURL(path string) string {
	return s.GetImageURL(path, "w1280")
}

// GetStillURL returns the still image URL
func (s *TMDBService) GetStillURL(path string) string {
	return s.GetImageURL(path, "w300")
}

// ParseDate parses a date string from TMDB
func ParseDate(dateStr string) (*time.Time, error) {
	if dateStr == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
