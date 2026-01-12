package repositories

import (
	"strings"
	"testing"

	"github.com/xc9973/go-tmdb-crawler/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestShowRepository_Search_SQLite(t *testing.T) {
	// Create in-memory database for each test
	db, err := gorm.Open(sqlite.Open("file:TestShowRepository_Search?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create tables
	err = db.AutoMigrate(&models.Show{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	// Insert test data
	shows := []models.Show{
		{TmdbID: 1, Name: "Breaking Bad", OriginalName: "Breaking Bad", Status: "Ended"},
		{TmdbID: 2, Name: "Game of Thrones", OriginalName: "Game of Thrones", Status: "Ended"},
		{TmdbID: 3, Name: "The Walking Dead", OriginalName: "The Walking Dead", Status: "Returning Series"},
		{TmdbID: 4, Name: "Stranger Things", OriginalName: "Stranger Things", Status: "Returning Series"},
		{TmdbID: 5, Name: "Breaking News", OriginalName: "Breaking News", Status: "Ended"},
	}

	for _, show := range shows {
		if err := db.Create(&show).Error; err != nil {
			t.Fatalf("Failed to insert test show: %v", err)
		}
	}

	repo := NewShowRepository(db)

	tests := []struct {
		name          string
		query         string
		expectedCount int
	}{
		{
			name:          "Search for 'breaking' (case insensitive)",
			query:         "breaking",
			expectedCount: 2,
		},
		{
			name:          "Search for 'BREAKING' (uppercase)",
			query:         "BREAKING",
			expectedCount: 2,
		},
		{
			name:          "Search for 'thrones' (partial match)",
			query:         "thrones",
			expectedCount: 1,
		},
		{
			name:          "Search for 'walking' (matches one show)",
			query:         "walking",
			expectedCount: 1,
		},
		{
			name:          "Search for 'xyz' (no match)",
			query:         "xyz",
			expectedCount: 0,
		},
		{
			name:          "Empty search string",
			query:         "",
			expectedCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shows, total, err := repo.Search(tt.query, 1, 10)
			if err != nil {
				t.Errorf("Search() error = %v", err)
				return
			}

			if total != int64(tt.expectedCount) {
				t.Errorf("Search() total = %d, want %d", total, tt.expectedCount)
			}

			if len(shows) != tt.expectedCount {
				t.Errorf("Search() returned %d shows, want %d", len(shows), tt.expectedCount)
			}

			// Verify all results contain the search term (case insensitive)
			if tt.query != "" && tt.expectedCount > 0 {
				for _, show := range shows {
					nameLower := strings.ToLower(show.Name)
					originalNameLower := strings.ToLower(show.OriginalName)
					queryLower := strings.ToLower(tt.query)
					if !strings.Contains(nameLower, queryLower) && !strings.Contains(originalNameLower, queryLower) {
						t.Errorf("Search() returned show '%s' which doesn't contain query '%s'", show.Name, tt.query)
					}
				}
			}
		})
	}
}

func TestShowRepository_ListFiltered_SQLite(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:TestShowRepository_ListFiltered?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	err = db.AutoMigrate(&models.Show{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	shows := []models.Show{
		{TmdbID: 1, Name: "Breaking Bad", OriginalName: "Breaking Bad", Status: "Ended"},
		{TmdbID: 2, Name: "Game of Thrones", OriginalName: "Game of Thrones", Status: "Ended"},
		{TmdbID: 3, Name: "The Walking Dead", OriginalName: "The Walking Dead", Status: "Returning Series"},
		{TmdbID: 4, Name: "Stranger Things", OriginalName: "Stranger Things", Status: "Returning Series"},
	}

	for _, show := range shows {
		if err := db.Create(&show).Error; err != nil {
			t.Fatalf("Failed to insert test show: %v", err)
		}
	}

	repo := NewShowRepository(db)

	tests := []struct {
		name          string
		status        string
		search        string
		expectedCount int
	}{
		{
			name:          "Filter by status 'Returning Series'",
			status:        "Returning Series",
			search:        "",
			expectedCount: 2,
		},
		{
			name:          "Filter by status 'Ended'",
			status:        "Ended",
			search:        "",
			expectedCount: 2,
		},
		{
			name:          "Search without status filter",
			status:        "",
			search:        "walking",
			expectedCount: 1,
		},
		{
			name:          "No filters",
			status:        "",
			search:        "",
			expectedCount: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shows, total, err := repo.ListFiltered(tt.status, tt.search, 1, 10)
			if err != nil {
				t.Errorf("ListFiltered() error = %v", err)
				return
			}

			if total != int64(tt.expectedCount) {
				t.Errorf("ListFiltered() total = %d, want %d", total, tt.expectedCount)
			}

			if len(shows) != tt.expectedCount {
				t.Errorf("ListFiltered() returned %d shows, want %d", len(shows), tt.expectedCount)
			}
		})
	}
}

func TestShowRepository_SearchCaseInsensitivity(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:TestShowRepository_Case?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	err = db.AutoMigrate(&models.Show{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	shows := []models.Show{
		{TmdbID: 1, Name: "Breaking Bad", OriginalName: "Breaking Bad", Status: "Ended"},
		{TmdbID: 2, Name: "Game of Thrones", OriginalName: "Game of Thrones", Status: "Ended"},
	}

	for _, show := range shows {
		if err := db.Create(&show).Error; err != nil {
			t.Fatalf("Failed to insert test show: %v", err)
		}
	}

	repo := NewShowRepository(db)

	// Test various case combinations
	searchTerms := []string{
		"breaking",
		"BREAKING",
		"Breaking",
		"BrEaKiNg",
	}

	for _, term := range searchTerms {
		t.Run("Search for '"+term+"'", func(t *testing.T) {
			shows, total, err := repo.Search(term, 1, 10)
			if err != nil {
				t.Errorf("Search() error = %v", err)
				return
			}

			// All case variations should return the same results
			if total != 1 {
				t.Errorf("Search(%q) total = %d, want 1", term, total)
			}

			if len(shows) != 1 {
				t.Errorf("Search(%q) returned %d shows, want 1", term, len(shows))
			}
		})
	}
}
