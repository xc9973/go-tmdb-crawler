-- TMDB Crawler Correction Feature Migration
-- Version: 006
-- Created: 2026-01-23

-- Add correction fields to shows table
ALTER TABLE shows ADD COLUMN refresh_threshold INTEGER DEFAULT NULL;
ALTER TABLE shows ADD COLUMN stale_detected_at TIMESTAMP DEFAULT NULL;
ALTER TABLE shows ADD COLUMN last_correction_result VARCHAR(50) DEFAULT NULL;

-- Create indexes for correction queries
CREATE INDEX idx_shows_stale_detected_at ON shows(stale_detected_at);
CREATE INDEX idx_shows_refresh_threshold ON shows(refresh_threshold);

-- Comments
COMMENT ON COLUMN shows.refresh_threshold IS 'Custom refresh threshold in days, NULL means auto-calculate';
COMMENT ON COLUMN shows.stale_detected_at IS 'Timestamp when the show was last detected as stale';
COMMENT ON COLUMN shows.last_correction_result IS 'Result of last correction attempt: pending/success/failed';
