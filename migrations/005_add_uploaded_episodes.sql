-- migrations/005_add_uploaded_episodes.sql
-- Episode upload tracking table
-- Purpose: Track which episodes have been uploaded to NAS to avoid duplicate uploads

CREATE TABLE IF NOT EXISTS uploaded_episodes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    episode_id INTEGER NOT NULL UNIQUE,
    uploaded BOOLEAN NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (episode_id) REFERENCES episodes(id) ON DELETE CASCADE
);

-- Index for faster lookups
CREATE INDEX IF NOT EXISTS idx_uploaded_episodes_episode_id ON uploaded_episodes(episode_id);
