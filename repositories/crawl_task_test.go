package repositories

import (
	"fmt"
	"testing"
	"time"

	"github.com/xc9973/go-tmdb-crawler/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupCrawlTaskDB(t *testing.T) *gorm.DB {
	// Use unique database name for each test to avoid conflicts
	dbName := fmt.Sprintf("file:CrawlTaskTest_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	err = db.AutoMigrate(&models.CrawlTask{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestCrawlTaskRepository_Create(t *testing.T) {
	db := setupCrawlTaskDB(t)
	repo := NewCrawlTaskRepository(db)

	task := &models.CrawlTask{
		Type:   "refresh_all",
		Status: "queued",
	}

	err := repo.Create(task)
	if err != nil {
		t.Errorf("Failed to create task: %v", err)
	}

	if task.ID == 0 {
		t.Error("Task ID should be set after creation")
	}
}

func TestCrawlTaskRepository_Update(t *testing.T) {
	db := setupCrawlTaskDB(t)
	repo := NewCrawlTaskRepository(db)

	task := &models.CrawlTask{
		Type:   "refresh_all",
		Status: "queued",
	}
	db.Create(task)

	task.Status = "running"
	now := time.Now()
	task.StartedAt = &now

	err := repo.Update(task)
	if err != nil {
		t.Errorf("Failed to update task: %v", err)
	}

	// Verify update
	var updated models.CrawlTask
	db.First(&updated, task.ID)
	if updated.Status != "running" {
		t.Errorf("Expected status 'running', got '%s'", updated.Status)
	}
}

func TestCrawlTaskRepository_GetByID(t *testing.T) {
	db := setupCrawlTaskDB(t)
	repo := NewCrawlTaskRepository(db)

	task := &models.CrawlTask{
		Type:   "crawl_by_id",
		Status: "success",
	}
	db.Create(task)

	found, err := repo.GetByID(task.ID)
	if err != nil {
		t.Errorf("Failed to get task by ID: %v", err)
	}

	if found.Type != task.Type {
		t.Errorf("Expected type %s, got %s", task.Type, found.Type)
	}
}

func TestCrawlTaskRepository_GetByStatus(t *testing.T) {
	db := setupCrawlTaskDB(t)
	repo := NewCrawlTaskRepository(db)

	// Create tasks with different statuses
	tasks := []*models.CrawlTask{
		{Type: "refresh_all", Status: "queued"},
		{Type: "crawl_by_id", Status: "queued"},
		{Type: "crawl_by_status", Status: "running"},
		{Type: "daily_job", Status: "success"},
	}
	for _, task := range tasks {
		db.Create(task)
	}

	// Test getting queued tasks
	found, total, err := repo.GetByStatus("queued", 1, 10)
	if err != nil {
		t.Errorf("Failed to get tasks by status: %v", err)
	}

	if total != 2 {
		t.Errorf("Expected total of 2, got %d", total)
	}

	if len(found) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(found))
	}
}

func TestCrawlTaskRepository_GetByStatus_Pagination(t *testing.T) {
	db := setupCrawlTaskDB(t)
	repo := NewCrawlTaskRepository(db)

	// Create 15 tasks
	for i := 0; i < 15; i++ {
		task := &models.CrawlTask{
			Type:   "refresh_all",
			Status: "queued",
		}
		db.Create(task)
	}

	// Test first page
	page1, total, err := repo.GetByStatus("queued", 1, 10)
	if err != nil {
		t.Errorf("Failed to get first page: %v", err)
	}

	if total != 15 {
		t.Errorf("Expected total of 15, got %d", total)
	}

	if len(page1) != 10 {
		t.Errorf("Expected 10 tasks on first page, got %d", len(page1))
	}

	// Test second page
	page2, _, err := repo.GetByStatus("queued", 2, 10)
	if err != nil {
		t.Errorf("Failed to get second page: %v", err)
	}

	if len(page2) != 5 {
		t.Errorf("Expected 5 tasks on second page, got %d", len(page2))
	}
}

func TestCrawlTaskRepository_GetRecent(t *testing.T) {
	db := setupCrawlTaskDB(t)
	repo := NewCrawlTaskRepository(db)

	// Create tasks with different timestamps
	for i := 0; i < 5; i++ {
		task := &models.CrawlTask{
			Type:   "refresh_all",
			Status: "success",
		}
		db.Create(task)
		// Small delay to ensure different timestamps
		time.Sleep(10 * time.Millisecond)
	}

	found, err := repo.GetRecent(3)
	if err != nil {
		t.Errorf("Failed to get recent tasks: %v", err)
	}

	if len(found) != 3 {
		t.Errorf("Expected 3 recent tasks, got %d", len(found))
	}
}

func TestCrawlTaskRepository_GetRunning(t *testing.T) {
	db := setupCrawlTaskDB(t)
	repo := NewCrawlTaskRepository(db)

	now := time.Now()

	// Create running tasks
	for i := 0; i < 3; i++ {
		task := &models.CrawlTask{
			Type:      "refresh_all",
			Status:    "running",
			StartedAt: &now,
		}
		db.Create(task)
	}

	// Create non-running tasks
	task := &models.CrawlTask{
		Type:   "refresh_all",
		Status: "queued",
	}
	db.Create(task)

	found, err := repo.GetRunning()
	if err != nil {
		t.Errorf("Failed to get running tasks: %v", err)
	}

	if len(found) != 3 {
		t.Errorf("Expected 3 running tasks, got %d", len(found))
	}
}

func TestCrawlTaskRepository_Delete(t *testing.T) {
	db := setupCrawlTaskDB(t)
	repo := NewCrawlTaskRepository(db)

	task := &models.CrawlTask{
		Type:   "refresh_all",
		Status: "queued",
	}
	db.Create(task)

	err := repo.Delete(task.ID)
	if err != nil {
		t.Errorf("Failed to delete task: %v", err)
	}

	// Verify deletion
	var count int64
	db.Model(&models.CrawlTask{}).Where("id = ?", task.ID).Count(&count)
	if count != 0 {
		t.Error("Task should be deleted")
	}
}

func TestCrawlTaskRepository_DeleteOld(t *testing.T) {
	db := setupCrawlTaskDB(t)
	repo := NewCrawlTaskRepository(db)

	// Create old task (30 days ago)
	oldTime := time.Now().AddDate(0, 0, -30)
	oldTask := &models.CrawlTask{
		Type:      "refresh_all",
		Status:    "success",
		CreatedAt: oldTime,
	}
	db.Create(oldTask)

	// Create recent task
	recentTask := &models.CrawlTask{
		Type:   "refresh_all",
		Status: "queued",
	}
	db.Create(recentTask)

	// Delete tasks older than 7 days
	err := repo.DeleteOld(7)
	if err != nil {
		t.Errorf("Failed to delete old tasks: %v", err)
	}

	// Verify old task is deleted
	var count int64
	db.Model(&models.CrawlTask{}).Where("id = ?", oldTask.ID).Count(&count)
	if count != 0 {
		t.Error("Old task should be deleted")
	}

	// Verify recent task still exists
	db.Model(&models.CrawlTask{}).Where("id = ?", recentTask.ID).Count(&count)
	if count != 1 {
		t.Error("Recent task should still exist")
	}
}

func TestCrawlTaskRepository_Count(t *testing.T) {
	db := setupCrawlTaskDB(t)
	repo := NewCrawlTaskRepository(db)

	// Create tasks
	for i := 0; i < 10; i++ {
		task := &models.CrawlTask{
			Type:   "refresh_all",
			Status: "queued",
		}
		db.Create(task)
	}

	count, err := repo.Count()
	if err != nil {
		t.Errorf("Failed to count tasks: %v", err)
	}

	if count != 10 {
		t.Errorf("Expected count of 10, got %d", count)
	}
}

func TestCrawlTaskRepository_CountByStatus(t *testing.T) {
	db := setupCrawlTaskDB(t)
	repo := NewCrawlTaskRepository(db)

	// Create tasks with different statuses
	tasks := []*models.CrawlTask{
		{Type: "refresh_all", Status: "queued"},
		{Type: "crawl_by_id", Status: "queued"},
		{Type: "crawl_by_status", Status: "running"},
		{Type: "daily_job", Status: "success"},
		{Type: "refresh_all", Status: "failed"},
	}
	for _, task := range tasks {
		db.Create(task)
	}

	// Count queued tasks
	queuedCount, err := repo.CountByStatus("queued")
	if err != nil {
		t.Errorf("Failed to count queued tasks: %v", err)
	}

	if queuedCount != 2 {
		t.Errorf("Expected 2 queued tasks, got %d", queuedCount)
	}

	// Count running tasks
	runningCount, err := repo.CountByStatus("running")
	if err != nil {
		t.Errorf("Failed to count running tasks: %v", err)
	}

	if runningCount != 1 {
		t.Errorf("Expected 1 running task, got %d", runningCount)
	}
}
