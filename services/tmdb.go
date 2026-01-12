package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/parnurzeal/gorequest"
	"github.com/xc9973/go-tmdb-crawler/dto"
)

// TMDBService handles all TMDB API interactions
type TMDBService struct {
	apiKey  string
	baseURL string
	lang    string
	client  *gorequest.SuperAgent
}

// NewTMDBService creates a new TMDB service instance
func NewTMDBService(apiKey, baseURL, lang string) *TMDBService {
	return &TMDBService{
		apiKey:  apiKey,
		baseURL: baseURL,
		lang:    lang,
		client:  gorequest.New(),
	}
}

// MustTMDBService returns a TMDB service and panics if API key is missing.
func MustTMDBService(apiKey, baseURL, lang string) *TMDBService {
	if apiKey == "" {
		panic("TMDB_API_KEY is required")
	}
	return NewTMDBService(apiKey, baseURL, lang)
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

// makeRequest makes an HTTP request to TMDB API
func (s *TMDBService) makeRequest(url string, result interface{}, queryParams ...map[string]string) error {
	// Build request
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
	request.Timeout(10 * time.Second)

	// Send request
	resp, body, errs := request.EndBytes()
	if len(errs) > 0 {
		return fmt.Errorf("request failed: %v", errs[0])
	}

	// Check status code
	if resp.StatusCode != 200 {
		var errResp dto.TMDBErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil {
			return fmt.Errorf("TMDB API error: %s", errResp.StatusMessage)
		}
		return fmt.Errorf("HTTP error: %d - %s", resp.StatusCode, string(body))
	}

	// Parse response
	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	return nil
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
