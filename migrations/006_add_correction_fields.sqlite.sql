-- TMDB Crawler Correction Feature Migration (SQLite)
-- Version: 006
-- Created: 2026-01-23

-- Add correction fields to shows table
-- Note: These columns will be added when shows table is created via GORM
-- This migration is for documentation purposes

-- The following columns should be in the shows table:
-- refresh_threshold INTEGER - Custom refresh threshold in days, NULL means auto-calculate
-- stale_detected_at TIMESTAMP - Timestamp when the show was last detected as stale
-- last_correction_result VARCHAR(50) - Result of last correction attempt: pending/success/failed

-- Create indexes for correction queries (to be created after table exists)
-- CREATE INDEX IF NOT EXISTS idx_shows_stale_detected_at ON shows(stale_detected_at);
-- CREATE INDEX IF NOT EXISTS idx_shows_refresh_threshold ON shows(refresh_threshold);

-- Note: SQLite doesn't support COMMENT ON COLUMN syntax
-- Comments are documented here for reference:
-- refresh_threshold: Custom refresh threshold in days, NULL means auto-calculate
-- stale_detected_at: Timestamp when the show was last detected as stale
-- last_correction_result: Result of last correction attempt: pending/success/failed
