package repositories

import (
	"testing"

	"github.com/xc9973/go-tmdb-crawler/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupUploadedEpisodeDB(t *testing.T) *gorm.DB {
	dbName := "file:UploadedEpisodeTest_?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	err = db.AutoMigrate(&models.UploadedEpisode{}, &models.Episode{}, &models.Show{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestUploadedEpisodeRepository_MarkUploaded(t *testing.T) {
	db := setupUploadedEpisodeDB(t)
	repo := NewUploadedEpisodeRepository(db)

	// Create test episode
	show := &models.Show{TmdbID: 1, Name: "Test Show"}
	db.Create(show)
	episode := &models.Episode{ShowID: show.ID, SeasonNumber: 1, EpisodeNumber: 1}
	db.Create(episode)

	// Test marking as uploaded
	err := repo.MarkUploaded(episode.ID)
	if err != nil {
		t.Fatalf("MarkUploaded() error = %v", err)
	}

	// Verify
	var ue models.UploadedEpisode
	err = db.Where("episode_id = ?", episode.ID).First(&ue).Error
	if err != nil {
		t.Fatalf("Failed to find uploaded episode: %v", err)
	}

	if !ue.Uploaded {
		t.Error("Expected uploaded to be true")
	}
}

func TestUploadedEpisodeRepository_IsUploaded(t *testing.T) {
	db := setupUploadedEpisodeDB(t)
	repo := NewUploadedEpisodeRepository(db)

	// Create test episode (use different tmdb_id to avoid UNIQUE constraint)
	show := &models.Show{TmdbID: 2, Name: "Test Show 2"}
	db.Create(show)
	episode := &models.Episode{ShowID: show.ID, SeasonNumber: 1, EpisodeNumber: 1}
	db.Create(episode)

	// Not uploaded initially
	uploaded, err := repo.IsUploaded(episode.ID)
	if err != nil {
		t.Fatalf("IsUploaded() error = %v", err)
	}
	if uploaded {
		t.Error("Expected uploaded to be false initially")
	}

	// Mark as uploaded
	repo.MarkUploaded(episode.ID)

	// Now should be uploaded
	uploaded, err = repo.IsUploaded(episode.ID)
	if err != nil {
		t.Fatalf("IsUploaded() error = %v", err)
	}
	if !uploaded {
		t.Error("Expected uploaded to be true after marking")
	}
}
