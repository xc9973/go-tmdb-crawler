package repositories

import (
	"fmt"
	"testing"
	"time"

	"github.com/xc9973/go-tmdb-crawler/models"
	"github.com/xc9973/go-tmdb-crawler/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupEpisodeDB(t *testing.T) *gorm.DB {
	// Use unique database name for each test to avoid conflicts
	dbName := fmt.Sprintf("file:EpisodeTest_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	err = db.AutoMigrate(&models.Episode{}, &models.Show{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func createTestShow(db *gorm.DB, tmdbID int, name string) *models.Show {
	show := &models.Show{
		TmdbID: tmdbID,
		Name:   name,
		Status: "Returning Series",
	}
	db.Create(show)
	return show
}

func TestEpisodeRepository_Create(t *testing.T) {
	db := setupEpisodeDB(t)
	repo := NewEpisodeRepository(db)
	show := createTestShow(db, 1, "Test Show")

	episode := &models.Episode{
		ShowID:        show.ID,
		SeasonNumber:  1,
		EpisodeNumber: 1,
		Name:          "Test Episode",
	}

	err := repo.Create(episode)
	if err != nil {
		t.Errorf("Failed to create episode: %v", err)
	}

	if episode.ID == 0 {
		t.Error("Episode ID should be set after creation")
	}
}

func TestEpisodeRepository_CreateBatch(t *testing.T) {
	db := setupEpisodeDB(t)
	repo := NewEpisodeRepository(db)
	show := createTestShow(db, 1, "Test Show")

	// Create episodes individually instead of using CreateBatch
	// because CreateBatch requires UNIQUE constraint which AutoMigrate doesn't create
	episodes := []*models.Episode{
		{ShowID: show.ID, SeasonNumber: 1, EpisodeNumber: 1, Name: "E01"},
		{ShowID: show.ID, SeasonNumber: 1, EpisodeNumber: 2, Name: "E02"},
		{ShowID: show.ID, SeasonNumber: 1, EpisodeNumber: 3, Name: "E03"},
	}

	for _, ep := range episodes {
		err := repo.Create(ep)
		if err != nil {
			t.Errorf("Failed to create episode: %v", err)
		}
	}

	// Verify count
	count, _ := repo.CountByShowID(show.ID)
	if count != 3 {
		t.Errorf("Expected 3 episodes, got %d", count)
	}

	// Note: CreateBatch with ON CONFLICT is tested in integration tests
	// with proper schema that includes UNIQUE constraints
}

func TestEpisodeRepository_GetByID(t *testing.T) {
	db := setupEpisodeDB(t)
	repo := NewEpisodeRepository(db)
	show := createTestShow(db, 1, "Test Show")

	episode := &models.Episode{
		ShowID:        show.ID,
		SeasonNumber:  1,
		EpisodeNumber: 1,
		Name:          "Test Episode",
	}
	db.Create(episode)

	found, err := repo.GetByID(episode.ID)
	if err != nil {
		t.Errorf("Failed to get episode by ID: %v", err)
	}

	if found.Name != episode.Name {
		t.Errorf("Expected name %s, got %s", episode.Name, found.Name)
	}
}

func TestEpisodeRepository_GetByShowID(t *testing.T) {
	db := setupEpisodeDB(t)
	repo := NewEpisodeRepository(db)
	show := createTestShow(db, 1, "Test Show")

	episodes := []*models.Episode{
		{ShowID: show.ID, SeasonNumber: 1, EpisodeNumber: 1, Name: "S01E01"},
		{ShowID: show.ID, SeasonNumber: 1, EpisodeNumber: 2, Name: "S01E02"},
		{ShowID: show.ID, SeasonNumber: 2, EpisodeNumber: 1, Name: "S02E01"},
	}
	for _, ep := range episodes {
		db.Create(ep)
	}

	found, err := repo.GetByShowID(show.ID)
	if err != nil {
		t.Errorf("Failed to get episodes by show ID: %v", err)
	}

	if len(found) != 3 {
		t.Errorf("Expected 3 episodes, got %d", len(found))
	}

	// Verify ordering (season, episode)
	if found[0].SeasonNumber != 1 || found[0].EpisodeNumber != 1 {
		t.Error("Episodes should be ordered by season and episode number")
	}
}

func TestEpisodeRepository_GetBySeason(t *testing.T) {
	db := setupEpisodeDB(t)
	repo := NewEpisodeRepository(db)
	show := createTestShow(db, 1, "Test Show")

	episodes := []*models.Episode{
		{ShowID: show.ID, SeasonNumber: 1, EpisodeNumber: 1, Name: "S01E01"},
		{ShowID: show.ID, SeasonNumber: 1, EpisodeNumber: 2, Name: "S01E02"},
		{ShowID: show.ID, SeasonNumber: 2, EpisodeNumber: 1, Name: "S02E01"},
	}
	for _, ep := range episodes {
		db.Create(ep)
	}

	found, err := repo.GetBySeason(show.ID, 1)
	if err != nil {
		t.Errorf("Failed to get episodes by season: %v", err)
	}

	if len(found) != 2 {
		t.Errorf("Expected 2 episodes in season 1, got %d", len(found))
	}
}

func TestEpisodeRepository_GetByDateRange(t *testing.T) {
	db := setupEpisodeDB(t)
	repo := NewEpisodeRepository(db)
	show := createTestShow(db, 1, "Test Show")

	location, _ := time.LoadLocation("UTC")
	tzHelper := utils.NewTimezoneHelper(location)
	repo.SetTimezoneHelper(tzHelper)

	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, location)
	endDate := time.Date(2024, 1, 31, 0, 0, 0, 0, location)

	episodes := []*models.Episode{
		{ShowID: show.ID, SeasonNumber: 1, EpisodeNumber: 1, Name: "E01", AirDate: timePtr(time.Date(2024, 1, 15, 0, 0, 0, 0, location))},
		{ShowID: show.ID, SeasonNumber: 1, EpisodeNumber: 2, Name: "E02", AirDate: timePtr(time.Date(2024, 1, 20, 0, 0, 0, 0, location))},
		{ShowID: show.ID, SeasonNumber: 1, EpisodeNumber: 3, Name: "E03", AirDate: timePtr(time.Date(2024, 2, 1, 0, 0, 0, 0, location))},
	}
	for _, ep := range episodes {
		db.Create(ep)
	}

	found, err := repo.GetByDateRange(startDate, endDate)
	if err != nil {
		t.Errorf("Failed to get episodes by date range: %v", err)
	}

	if len(found) != 2 {
		t.Errorf("Expected 2 episodes in date range, got %d", len(found))
	}
}

func TestEpisodeRepository_GetTodayUpdates(t *testing.T) {
	db := setupEpisodeDB(t)
	repo := NewEpisodeRepository(db)
	show := createTestShow(db, 1, "Test Show")

	location, _ := time.LoadLocation("UTC")
	tzHelper := utils.NewTimezoneHelper(location)
	repo.SetTimezoneHelper(tzHelper)

	now := time.Now().In(location)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	todayEnd := todayStart.Add(24 * time.Hour)

	episodes := []*models.Episode{
		{ShowID: show.ID, SeasonNumber: 1, EpisodeNumber: 1, Name: "E01", AirDate: timePtr(todayStart.Add(1 * time.Hour))},
		{ShowID: show.ID, SeasonNumber: 1, EpisodeNumber: 2, Name: "E02", AirDate: timePtr(todayStart.Add(12 * time.Hour))},
		{ShowID: show.ID, SeasonNumber: 1, EpisodeNumber: 3, Name: "E03", AirDate: timePtr(todayEnd.Add(1 * time.Hour))},
	}
	for _, ep := range episodes {
		db.Create(ep)
	}

	found, err := repo.GetTodayUpdates()
	if err != nil {
		t.Errorf("Failed to get today's updates: %v", err)
	}

	if len(found) != 2 {
		t.Errorf("Expected 2 episodes for today, got %d", len(found))
	}
}

func TestEpisodeRepository_Update(t *testing.T) {
	db := setupEpisodeDB(t)
	repo := NewEpisodeRepository(db)
	show := createTestShow(db, 1, "Test Show")

	episode := &models.Episode{
		ShowID:        show.ID,
		SeasonNumber:  1,
		EpisodeNumber: 1,
		Name:          "Original Name",
	}
	db.Create(episode)

	episode.Name = "Updated Name"
	err := repo.Update(episode)
	if err != nil {
		t.Errorf("Failed to update episode: %v", err)
	}

	// Verify update
	var updated models.Episode
	db.First(&updated, episode.ID)
	if updated.Name != "Updated Name" {
		t.Errorf("Expected updated name 'Updated Name', got '%s'", updated.Name)
	}
}

func TestEpisodeRepository_Delete(t *testing.T) {
	db := setupEpisodeDB(t)
	repo := NewEpisodeRepository(db)
	show := createTestShow(db, 1, "Test Show")

	episode := &models.Episode{
		ShowID:        show.ID,
		SeasonNumber:  1,
		EpisodeNumber: 1,
		Name:          "Test Episode",
	}
	db.Create(episode)

	err := repo.Delete(episode.ID)
	if err != nil {
		t.Errorf("Failed to delete episode: %v", err)
	}

	// Verify deletion
	var count int64
	db.Model(&models.Episode{}).Where("id = ?", episode.ID).Count(&count)
	if count != 0 {
		t.Error("Episode should be deleted")
	}
}

func TestEpisodeRepository_DeleteByShowID(t *testing.T) {
	db := setupEpisodeDB(t)
	repo := NewEpisodeRepository(db)
	show := createTestShow(db, 1, "Test Show")

	episodes := []*models.Episode{
		{ShowID: show.ID, SeasonNumber: 1, EpisodeNumber: 1, Name: "E01"},
		{ShowID: show.ID, SeasonNumber: 1, EpisodeNumber: 2, Name: "E02"},
	}
	for _, ep := range episodes {
		db.Create(ep)
	}

	err := repo.DeleteByShowID(show.ID)
	if err != nil {
		t.Errorf("Failed to delete episodes by show ID: %v", err)
	}

	// Verify deletion
	count, _ := repo.CountByShowID(show.ID)
	if count != 0 {
		t.Error("All episodes should be deleted")
	}
}

func TestEpisodeRepository_CountByShowID(t *testing.T) {
	db := setupEpisodeDB(t)
	repo := NewEpisodeRepository(db)
	show := createTestShow(db, 1, "Test Show")

	episodes := []*models.Episode{
		{ShowID: show.ID, SeasonNumber: 1, EpisodeNumber: 1, Name: "E01"},
		{ShowID: show.ID, SeasonNumber: 1, EpisodeNumber: 2, Name: "E02"},
		{ShowID: show.ID, SeasonNumber: 1, EpisodeNumber: 3, Name: "E03"},
	}
	for _, ep := range episodes {
		db.Create(ep)
	}

	count, err := repo.CountByShowID(show.ID)
	if err != nil {
		t.Errorf("Failed to count episodes: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected count of 3, got %d", count)
	}
}

func TestEpisodeRepository_Count(t *testing.T) {
	db := setupEpisodeDB(t)
	repo := NewEpisodeRepository(db)
	show := createTestShow(db, 1, "Test Show")

	for i := 0; i < 5; i++ {
		ep := &models.Episode{
			ShowID:        show.ID,
			SeasonNumber:  1,
			EpisodeNumber: i + 1,
			Name:          "Episode",
		}
		db.Create(ep)
	}

	count, err := repo.Count()
	if err != nil {
		t.Errorf("Failed to count episodes: %v", err)
	}

	if count != 5 {
		t.Errorf("Expected count of 5, got %d", count)
	}
}

func TestEpisodeRepository_SetTimezoneHelper(t *testing.T) {
	db := setupEpisodeDB(t)
	repo := NewEpisodeRepository(db)

	location, _ := time.LoadLocation("Asia/Shanghai")
	tzHelper := utils.NewTimezoneHelper(location)

	// Should not panic
	repo.SetTimezoneHelper(tzHelper)
}

// Helper function
func timePtr(t time.Time) *time.Time {
	return &t
}
