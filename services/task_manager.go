package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/xc9973/go-tmdb-crawler/models"
	"github.com/xc9973/go-tmdb-crawler/repositories"
)

const (
	taskStatusQueued  = "queued"
	taskStatusRunning = "running"
	taskStatusSuccess = "success"
	taskStatusFailed  = "failed"
)

// TaskManager manages async crawl tasks and persists status.
type TaskManager struct {
	tasks   repositories.CrawlTaskRepository
	crawler *CrawlerService
}

// NewTaskManager creates a task manager instance.
func NewTaskManager(tasks repositories.CrawlTaskRepository, crawler *CrawlerService) *TaskManager {
	return &TaskManager{
		tasks:   tasks,
		crawler: crawler,
	}
}

// StartRefreshAll starts a refresh-all task in background.
func (m *TaskManager) StartRefreshAll() (*models.CrawlTask, error) {
	return m.startTask("refresh_all", nil, func() error {
		return m.crawler.RefreshAll()
	})
}

// StartCrawlByStatus starts a status crawl task in background.
func (m *TaskManager) StartCrawlByStatus(status string) (*models.CrawlTask, error) {
	params := map[string]string{"status": status}
	return m.startTask("crawl_by_status", params, func() error {
		return m.crawler.CrawlByStatus(status)
	})
}

// GetTask returns task status by id.
func (m *TaskManager) GetTask(id uint) (*models.CrawlTask, error) {
	return m.tasks.GetByID(id)
}

func (m *TaskManager) startTask(taskType string, params map[string]string, runner func() error) (*models.CrawlTask, error) {
	paramsJSON := ""
	if params != nil {
		b, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal task params: %w", err)
		}
		paramsJSON = string(b)
	}

	now := time.Now()
	task := &models.CrawlTask{
		Type:      taskType,
		Status:    taskStatusQueued,
		Params:    paramsJSON,
		CreatedAt: now,
	}

	if err := m.tasks.Create(task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	go func(t *models.CrawlTask) {
		startedAt := time.Now()
		t.Status = taskStatusRunning
		t.StartedAt = &startedAt
		_ = m.tasks.Update(t)

		err := runner()
		finishedAt := time.Now()
		t.FinishedAt = &finishedAt
		if err != nil {
			t.Status = taskStatusFailed
			t.ErrorMessage = err.Error()
		} else {
			t.Status = taskStatusSuccess
		}
		_ = m.tasks.Update(t)
	}(task)

	return task, nil
}
