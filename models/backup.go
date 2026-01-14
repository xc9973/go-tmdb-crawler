package models

import "time"

// BackupVersion is the current backup format version
const BackupVersion = "1.0"

// BackupExport represents the complete backup export structure
type BackupExport struct {
	Version    string      `json:"version"`     // e.g. "1.0"
	ExportedAt time.Time   `json:"exported_at"`
	AppVersion string      `json:"app_version"`
	Stats      BackupStats `json:"stats"`
	Data       BackupData  `json:"data"`
}

// BackupStats holds statistics about the backup
type BackupStats struct {
	Shows          int `json:"shows"`
	Episodes       int `json:"episodes"`
	CrawlLogs      int `json:"crawl_logs"`
	TelegraphPosts int `json:"telegraph_posts"`
}

// BackupData holds all the data tables
type BackupData struct {
	Shows          []Show          `json:"shows"`
	Episodes       []Episode       `json:"episodes"`
	CrawlLogs      []CrawlLog      `json:"crawl_logs"`
	TelegraphPosts []TelegraphPost `json:"telegraph_posts"`
}

// BackupStatus represents the current backup status
type BackupStatus struct {
	LastBackup *time.Time  `json:"last_backup,omitempty"`
	Stats      BackupStats `json:"stats"`
}

// ImportResult represents the result of an import operation
type ImportResult struct {
	ShowsImported        int `json:"shows_imported"`
	EpisodesImported     int `json:"episodes_imported"`
	CrawlLogsImported    int `json:"crawl_logs_imported"`
	TelegraphPostsImported int `json:"telegraph_posts_imported"`
	ConflictsSkipped     int `json:"conflicts_skipped"`
}
