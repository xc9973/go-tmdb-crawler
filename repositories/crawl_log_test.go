package repositories

import (
	"fmt"
	"testing"
	"time"

	"github.com/xc9973/go-tmdb-crawler/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupCrawlLogDB(t *testing.T) *gorm.DB {
	// Use unique database name for each test to avoid conflicts
	dbName := fmt.Sprintf("file:CrawlLogTest_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	err = db.AutoMigrate(&models.CrawlLog{}, &models.Show{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func createTestShowForLog(db *gorm.DB, tmdbID int, name string) *models.Show {
	show := &models.Show{
		TmdbID: tmdbID,
		Name:   name,
		Status: "Returning Series",
	}
	db.Create(show)
	return show
}

func TestCrawlLogRepository_Create(t *testing.T) {
	db := setupCrawlLogDB(t)
	repo := NewCrawlLogRepository(db)

	log := &models.CrawlLog{
		TmdbID:        123,
		Action:        "fetch",
		Status:        "success",
		EpisodesCount: 10,
		DurationMs:    1000,
	}

	err := repo.Create(log)
	if err != nil {
		t.Errorf("Failed to create log: %v", err)
	}

	if log.ID == 0 {
		t.Error("Log ID should be set after creation")
	}
}

func TestCrawlLogRepository_GetByID(t *testing.T) {
	db := setupCrawlLogDB(t)
	repo := NewCrawlLogRepository(db)

	log := &models.CrawlLog{
		TmdbID:        123,
		Action:        "fetch",
		Status:        "success",
		EpisodesCount: 10,
	}
	db.Create(log)

	found, err := repo.GetByID(log.ID)
	if err != nil {
		t.Errorf("Failed to get log by ID: %v", err)
	}

	if found.Action != log.Action {
		t.Errorf("Expected action %s, got %s", log.Action, found.Action)
	}
}

func TestCrawlLogRepository_GetByShowID(t *testing.T) {
	db := setupCrawlLogDB(t)
	repo := NewCrawlLogRepository(db)
	show := createTestShowForLog(db, 1, "Test Show")

	// Create logs for the show
	for i := 0; i < 5; i++ {
		log := &models.CrawlLog{
			ShowID:        &show.ID,
			TmdbID:        123,
			Action:        "fetch",
			Status:        "success",
			EpisodesCount: 10,
		}
		db.Create(log)
	}

	// Create log for different show
	show2 := createTestShowForLog(db, 2, "Test Show 2")
	log2 := &models.CrawlLog{
		ShowID:        &show2.ID,
		TmdbID:        456,
		Action:        "fetch",
		Status:        "success",
		EpisodesCount: 5,
	}
	db.Create(log2)

	// Get logs for first show with limit
	found, err := repo.GetByShowID(show.ID, 3)
	if err != nil {
		t.Errorf("Failed to get logs by show ID: %v", err)
	}

	if len(found) != 3 {
		t.Errorf("Expected 3 logs (limited), got %d", len(found))
	}

	// Verify all logs belong to the show
	for _, log := range found {
		if log.ShowID == nil || *log.ShowID != show.ID {
			t.Error("Log should belong to the show")
		}
	}
}

func TestCrawlLogRepository_GetByShowID_NoLimit(t *testing.T) {
	db := setupCrawlLogDB(t)
	repo := NewCrawlLogRepository(db)
	show := createTestShowForLog(db, 1, "Test Show")

	// Create logs
	for i := 0; i < 5; i++ {
		log := &models.CrawlLog{
			ShowID:        &show.ID,
			TmdbID:        123,
			Action:        "fetch",
			Status:        "success",
			EpisodesCount: 10,
		}
		db.Create(log)
	}

	// Get all logs (no limit)
	found, err := repo.GetByShowID(show.ID, 0)
	if err != nil {
		t.Errorf("Failed to get logs by show ID: %v", err)
	}

	if len(found) != 5 {
		t.Errorf("Expected 5 logs (no limit), got %d", len(found))
	}
}

func TestCrawlLogRepository_GetRecent(t *testing.T) {
	db := setupCrawlLogDB(t)
	repo := NewCrawlLogRepository(db)

	// Create logs
	for i := 0; i < 10; i++ {
		log := &models.CrawlLog{
			TmdbID:        100 + i,
			Action:        "fetch",
			Status:        "success",
			EpisodesCount: 10,
		}
		db.Create(log)
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	found, err := repo.GetRecent(5)
	if err != nil {
		t.Errorf("Failed to get recent logs: %v", err)
	}

	if len(found) != 5 {
		t.Errorf("Expected 5 recent logs, got %d", len(found))
	}
}

func TestCrawlLogRepository_GetByStatus(t *testing.T) {
	db := setupCrawlLogDB(t)
	repo := NewCrawlLogRepository(db)

	// Create logs with different statuses
	logs := []*models.CrawlLog{
		{TmdbID: 1, Action: "fetch", Status: "success"},
		{TmdbID: 2, Action: "fetch", Status: "success"},
		{TmdbID: 3, Action: "refresh", Status: "failed"},
		{TmdbID: 4, Action: "batch", Status: "partial"},
	}
	for _, log := range logs {
		db.Create(log)
	}

	// Get success logs
	found, total, err := repo.GetByStatus("success", 1, 10)
	if err != nil {
		t.Errorf("Failed to get logs by status: %v", err)
	}

	if total != 2 {
		t.Errorf("Expected total of 2, got %d", total)
	}

	if len(found) != 2 {
		t.Errorf("Expected 2 logs, got %d", len(found))
	}
}

func TestCrawlLogRepository_GetByStatus_Pagination(t *testing.T) {
	db := setupCrawlLogDB(t)
	repo := NewCrawlLogRepository(db)

	// Create 15 logs
	for i := 0; i < 15; i++ {
		log := &models.CrawlLog{
			TmdbID:        100 + i,
			Action:        "fetch",
			Status:        "success",
			EpisodesCount: 10,
		}
		db.Create(log)
	}

	// Test first page
	page1, total, err := repo.GetByStatus("success", 1, 10)
	if err != nil {
		t.Errorf("Failed to get first page: %v", err)
	}

	if total != 15 {
		t.Errorf("Expected total of 15, got %d", total)
	}

	if len(page1) != 10 {
		t.Errorf("Expected 10 logs on first page, got %d", len(page1))
	}

	// Test second page
	page2, _, err := repo.GetByStatus("success", 2, 10)
	if err != nil {
		t.Errorf("Failed to get second page: %v", err)
	}

	if len(page2) != 5 {
		t.Errorf("Expected 5 logs on second page, got %d", len(page2))
	}
}

func TestCrawlLogRepository_GetByDateRange(t *testing.T) {
	db := setupCrawlLogDB(t)
	repo := NewCrawlLogRepository(db)

	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

	// Create logs in different date ranges
	logs := []*models.CrawlLog{
		{TmdbID: 1, Action: "fetch", Status: "success", CreatedAt: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
		{TmdbID: 2, Action: "fetch", Status: "success", CreatedAt: time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)},
		{TmdbID: 3, Action: "fetch", Status: "success", CreatedAt: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)},
	}
	for _, log := range logs {
		db.Create(log)
	}

	found, err := repo.GetByDateRange(startDate, endDate)
	if err != nil {
		t.Errorf("Failed to get logs by date range: %v", err)
	}

	if len(found) != 2 {
		t.Errorf("Expected 2 logs in date range, got %d", len(found))
	}
}

func TestCrawlLogRepository_Delete(t *testing.T) {
	db := setupCrawlLogDB(t)
	repo := NewCrawlLogRepository(db)

	log := &models.CrawlLog{
		TmdbID:        123,
		Action:        "fetch",
		Status:        "success",
		EpisodesCount: 10,
	}
	db.Create(log)

	err := repo.Delete(log.ID)
	if err != nil {
		t.Errorf("Failed to delete log: %v", err)
	}

	// Verify deletion
	var count int64
	db.Model(&models.CrawlLog{}).Where("id = ?", log.ID).Count(&count)
	if count != 0 {
		t.Error("Log should be deleted")
	}
}

func TestCrawlLogRepository_DeleteOld(t *testing.T) {
	db := setupCrawlLogDB(t)
	repo := NewCrawlLogRepository(db)

	// Create old log (30 days ago)
	oldTime := time.Now().AddDate(0, 0, -30)
	oldLog := &models.CrawlLog{
		TmdbID:        1,
		Action:        "fetch",
		Status:        "success",
		EpisodesCount: 10,
		CreatedAt:     oldTime,
	}
	db.Create(oldLog)

	// Create recent log
	recentLog := &models.CrawlLog{
		TmdbID:        2,
		Action:        "fetch",
		Status:        "success",
		EpisodesCount: 10,
	}
	db.Create(recentLog)

	// Delete logs older than 7 days
	err := repo.DeleteOld(7)
	if err != nil {
		t.Errorf("Failed to delete old logs: %v", err)
	}

	// Verify old log is deleted
	var count int64
	db.Model(&models.CrawlLog{}).Where("id = ?", oldLog.ID).Count(&count)
	if count != 0 {
		t.Error("Old log should be deleted")
	}

	// Verify recent log still exists
	db.Model(&models.CrawlLog{}).Where("id = ?", recentLog.ID).Count(&count)
	if count != 1 {
		t.Error("Recent log should still exist")
	}
}

func TestCrawlLogRepository_Count(t *testing.T) {
	db := setupCrawlLogDB(t)
	repo := NewCrawlLogRepository(db)

	// Create logs
	for i := 0; i < 20; i++ {
		log := &models.CrawlLog{
			TmdbID:        100 + i,
			Action:        "fetch",
			Status:        "success",
			EpisodesCount: 10,
		}
		db.Create(log)
	}

	count, err := repo.Count()
	if err != nil {
		t.Errorf("Failed to count logs: %v", err)
	}

	if count != 20 {
		t.Errorf("Expected count of 20, got %d", count)
	}
}

func TestCrawlLogRepository_CountByStatus(t *testing.T) {
	db := setupCrawlLogDB(t)
	repo := NewCrawlLogRepository(db)

	// Create logs with different statuses
	logs := []*models.CrawlLog{
		{TmdbID: 1, Action: "fetch", Status: "success"},
		{TmdbID: 2, Action: "fetch", Status: "success"},
		{TmdbID: 3, Action: "fetch", Status: "success"},
		{TmdbID: 4, Action: "refresh", Status: "failed"},
		{TmdbID: 5, Action: "batch", Status: "partial"},
	}
	for _, log := range logs {
		db.Create(log)
	}

	// Count success logs
	successCount, err := repo.CountByStatus("success")
	if err != nil {
		t.Errorf("Failed to count success logs: %v", err)
	}

	if successCount != 3 {
		t.Errorf("Expected 3 success logs, got %d", successCount)
	}

	// Count failed logs
	failedCount, err := repo.CountByStatus("failed")
	if err != nil {
		t.Errorf("Failed to count failed logs: %v", err)
	}

	if failedCount != 1 {
		t.Errorf("Expected 1 failed log, got %d", failedCount)
	}
}
