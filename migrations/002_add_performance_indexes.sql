-- Performance Optimization Migration
-- Version: 2.0
-- Created: 2026-01-12
-- Description: Add additional indexes for improved query performance

-- ============================================
-- 1. Shows Table - Additional Indexes
-- ============================================

-- Composite index for status + created_at (common query pattern)
CREATE INDEX IF NOT EXISTS idx_shows_status_created_at ON shows(status, created_at DESC);

-- Index for name searches (ILIKE/LIKE operations)
CREATE INDEX IF NOT EXISTS idx_shows_name_trgm ON shows USING gin(name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_shows_original_name_trgm ON shows USING gin(original_name gin_trgm_ops);

-- Note: gin_trgm_ops requires pg_trgm extension
-- CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Composite index for upcoming shows (status + next_air_date)
CREATE INDEX IF NOT EXISTS idx_shows_status_next_air_date ON shows(status, next_air_date) 
    WHERE next_air_date IS NOT NULL;

-- Index for popularity sorting
CREATE INDEX IF NOT EXISTS idx_shows_popularity ON shows(popularity DESC) 
    WHERE popularity IS NOT NULL;

-- ============================================
-- 2. Episodes Table - Additional Indexes
-- ============================================

-- Composite index for show + season + episode (unique constraint already exists)
-- This index helps with JOIN operations
CREATE INDEX IF NOT EXISTS idx_episodes_show_season_episode ON episodes(show_id, season_number, episode_number);

-- Composite index for date range queries with show
CREATE INDEX IF NOT EXISTS idx_episodes_show_air_date ON episodes(show_id, air_date DESC);

-- Index for recent episodes (air_date + created_at)
CREATE INDEX IF NOT EXISTS idx_episodes_air_date_created ON episodes(air_date DESC, created_at DESC);

-- ============================================
-- 3. Crawl Logs Table - Additional Indexes
-- ============================================

-- Composite index for show logs with pagination
CREATE INDEX IF NOT EXISTS idx_crawl_logs_show_created ON crawl_logs(show_id, created_at DESC);

-- Composite index for status filtering with time
CREATE INDEX IF NOT EXISTS idx_crawl_logs_status_created ON crawl_logs(status, created_at DESC);

-- ============================================
-- 4. Telegraph Posts Table - Additional Indexes
-- ============================================

-- Index for title searches
CREATE INDEX IF NOT EXISTS idx_telegraph_posts_title_trgm ON telegraph_posts USING gin(title gin_trgm_ops);

-- ============================================
-- 5. Crawl Tasks Table - Additional Indexes
-- ============================================

-- Composite index for active tasks
CREATE INDEX IF NOT EXISTS idx_crawl_tasks_status_created ON crawl_tasks(status, created_at DESC);

-- Index for task type filtering
CREATE INDEX IF NOT EXISTS idx_crawl_tasks_type_status ON crawl_tasks(type, status);

-- ============================================
-- 6. Partial Indexes for Common Queries
-- ============================================

-- Partial index for returning series (most frequently queried)
CREATE INDEX IF NOT EXISTS idx_shows_returning_series ON shows(next_air_date ASC) 
    WHERE status = 'Returning Series' AND next_air_date IS NOT NULL;

-- Partial index for recent crawl logs (last 30 days)
CREATE INDEX IF NOT EXISTS idx_crawl_logs_recent ON crawl_logs(created_at DESC)
    WHERE created_at > NOW() - INTERVAL '30 days';

-- Partial index for today's episodes
CREATE INDEX IF NOT EXISTS idx_episodes_today ON episodes(air_date)
    WHERE air_date >= CURRENT_DATE;

-- ============================================
-- 7. Covering Indexes for Specific Queries
-- ============================================

-- Covering index for shows list query (includes commonly selected columns)
CREATE INDEX IF NOT EXISTS idx_shows_list_covering ON shows(created_at DESC) 
    INCLUDE (name, status, poster_path, next_air_date);

-- ============================================
-- 8. Statistics Update
-- ============================================

-- Update table statistics for better query planning
ANALYZE shows;
ANALYZE episodes;
ANALYZE crawl_logs;
ANALYZE telegraph_posts;
ANALYZE crawl_tasks;

-- ============================================
-- 9. Comments
-- ============================================

COMMENT ON INDEX idx_shows_status_created_at IS 'Composite index for status filtering with pagination';
COMMENT ON INDEX idx_shows_name_trgm IS 'GIN index for name text search using trigram';
COMMENT ON INDEX idx_shows_status_next_air_date IS 'Index for upcoming shows query';
COMMENT ON INDEX idx_episodes_show_season_episode IS 'Composite index for episode lookups';
COMMENT ON INDEX idx_shows_returning_series IS 'Partial index for returning series only';
COMMENT ON INDEX idx_crawl_logs_recent IS 'Partial index for recent crawl logs (30 days)';

-- ============================================
-- End of Performance Optimization Migration
-- ============================================
