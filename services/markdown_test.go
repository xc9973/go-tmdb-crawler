package services

import (
	"testing"
	"time"

	"github.com/xc9973/go-tmdb-crawler/models"
)

func TestMarkdownService_GenerateShowContent_SeasonOrder(t *testing.T) {
	show := &models.Show{
		TmdbID:       1,
		Name:         "Test Show",
		OriginalName: "Test Show",
		Status:       "Returning Series",
		Type:         "Scripted",
		Language:     "en",
		VoteAverage:  8.5,
		VoteCount:    1000,
		Overview:     "A test show for season ordering",
	}

	// Create episodes in random season order
	episodes := []*models.Episode{
		{
			ShowID:        1,
			SeasonNumber:  3,
			EpisodeNumber: 1,
			Name:          "S03E01",
			Overview:      "Episode 1 of season 3",
			AirDate:       timePtr(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
		},
		{
			ShowID:        1,
			SeasonNumber:  1,
			EpisodeNumber: 1,
			Name:          "S01E01",
			Overview:      "Episode 1 of season 1",
			AirDate:       timePtr(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
		},
		{
			ShowID:        1,
			SeasonNumber:  2,
			EpisodeNumber: 1,
			Name:          "S02E01",
			Overview:      "Episode 1 of season 2",
			AirDate:       timePtr(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
		},
		{
			ShowID:        1,
			SeasonNumber:  5,
			EpisodeNumber: 1,
			Name:          "S05E01",
			Overview:      "Episode 1 of season 5",
			AirDate:       timePtr(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
		},
		{
			ShowID:        1,
			SeasonNumber:  4,
			EpisodeNumber: 1,
			Name:          "S04E01",
			Overview:      "Episode 1 of season 4",
			AirDate:       timePtr(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
		},
	}

	// Set Show reference for each episode
	for _, ep := range episodes {
		ep.Show = show
	}

	markdownService := &MarkdownService{}
	markdown := markdownService.GenerateShowContent(show, episodes)

	// Verify seasons are in ascending order
	// Check that "第1季" appears before "第2季", etc.
	season1Pos := findSubstring(markdown, "第1季")
	season2Pos := findSubstring(markdown, "第2季")
	season3Pos := findSubstring(markdown, "第3季")
	season4Pos := findSubstring(markdown, "第4季")
	season5Pos := findSubstring(markdown, "第5季")

	if season1Pos == -1 || season2Pos == -1 || season3Pos == -1 || season4Pos == -1 || season5Pos == -1 {
		t.Fatal("Not all seasons found in markdown output")
	}

	// Verify order: season 1 < season 2 < season 3 < season 4 < season 5
	if !(season1Pos < season2Pos && season2Pos < season3Pos && season3Pos < season4Pos && season4Pos < season5Pos) {
		t.Error("Seasons are not in ascending order")
		t.Logf("Season positions: S1=%d, S2=%d, S3=%d, S4=%d, S5=%d",
			season1Pos, season2Pos, season3Pos, season4Pos, season5Pos)
	}
}

func TestMarkdownService_GenerateShowContent_SeasonOrderStability(t *testing.T) {
	show := &models.Show{
		TmdbID:       1,
		Name:         "Test Show",
		OriginalName: "Test Show",
		Status:       "Returning Series",
		Type:         "Scripted",
		Language:     "en",
		VoteAverage:  8.5,
		VoteCount:    1000,
	}

	// Create episodes
	episodes := []*models.Episode{
		{ShowID: 1, SeasonNumber: 3, EpisodeNumber: 1, Name: "S03E01"},
		{ShowID: 1, SeasonNumber: 1, EpisodeNumber: 1, Name: "S01E01"},
		{ShowID: 1, SeasonNumber: 2, EpisodeNumber: 1, Name: "S02E01"},
	}

	for _, ep := range episodes {
		ep.Show = show
	}

	markdownService := &MarkdownService{}

	// Generate markdown multiple times and verify the output is identical
	var outputs []string
	for i := 0; i < 5; i++ {
		markdown := markdownService.GenerateShowContent(show, episodes)
		outputs = append(outputs, markdown)
	}

	// Verify all outputs are identical
	for i := 1; i < len(outputs); i++ {
		if outputs[i] != outputs[0] {
			t.Errorf("Output %d differs from output 0", i)
			t.Logf("Output 0 length: %d", len(outputs[0]))
			t.Logf("Output %d length: %d", i, len(outputs[i]))
		}
	}
}

func TestMarkdownService_GenerateShowContent_SingleSeason(t *testing.T) {
	show := &models.Show{
		TmdbID:       1,
		Name:         "Single Season Show",
		OriginalName: "Single Season Show",
		Status:       "Ended",
		Type:         "Scripted",
		Language:     "en",
		VoteAverage:  9.0,
		VoteCount:    500,
	}

	episodes := []*models.Episode{
		{ShowID: 1, SeasonNumber: 1, EpisodeNumber: 1, Name: "Pilot"},
		{ShowID: 1, SeasonNumber: 1, EpisodeNumber: 2, Name: "Episode 2"},
		{ShowID: 1, SeasonNumber: 1, EpisodeNumber: 3, Name: "Episode 3"},
	}

	for _, ep := range episodes {
		ep.Show = show
	}

	markdownService := &MarkdownService{}
	markdown := markdownService.GenerateShowContent(show, episodes)

	// Verify only one season section exists
	seasonCount := countOccurrences(markdown, "### 第")
	if seasonCount != 1 {
		t.Errorf("Expected 1 season section, got %d", seasonCount)
	}

	// Verify it's season 1
	if !containsSubstring(markdown, "第1季") {
		t.Error("Season 1 not found in output")
	}
}

func TestMarkdownService_GenerateShowContent_NoEpisodes(t *testing.T) {
	show := &models.Show{
		TmdbID:       1,
		Name:         "Empty Show",
		OriginalName: "Empty Show",
		Status:       "Returning Series",
		Type:         "Scripted",
		Language:     "en",
		VoteAverage:  0.0,
		VoteCount:    0,
	}

	episodes := []*models.Episode{}

	markdownService := &MarkdownService{}
	markdown := markdownService.GenerateShowContent(show, episodes)

	// Verify no season sections exist
	seasonCount := countOccurrences(markdown, "### 第")
	if seasonCount != 0 {
		t.Errorf("Expected 0 season sections, got %d", seasonCount)
	}

	// Verify show info is still present
	if !containsSubstring(markdown, show.Name) {
		t.Error("Show name not found in output")
	}
}

// Helper functions

func timePtr(t time.Time) *time.Time {
	return &t
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func containsSubstring(s, substr string) bool {
	return findSubstring(s, substr) != -1
}

func countOccurrences(s, substr string) int {
	count := 0
	pos := 0
	for {
		idx := findSubstring(s[pos:], substr)
		if idx == -1 {
			break
		}
		count++
		pos += idx + len(substr)
	}
	return count
}
