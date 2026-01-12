-- TMDB Crawler Database Schema
-- Version: 1.0
-- Created: 2026-01-11

-- ============================================
-- 1. Shows Table (剧集表)
-- ============================================
CREATE TABLE IF NOT EXISTS shows (
    id SERIAL PRIMARY KEY,
    tmdb_id INTEGER UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    original_name VARCHAR(255),
    status VARCHAR(50),  -- 'Ended'/'Returning Series'/'Canceled'
    first_air_date DATE,
    overview TEXT,
    poster_path VARCHAR(512),
    backdrop_path VARCHAR(512),
    genres VARCHAR(255),  -- JSON array string
    popularity DECIMAL(5,2),
    vote_average DECIMAL(3,1),
    vote_count INTEGER,
    
    -- Local fields (本地字段)
    last_season_number INTEGER,     -- 最新季数
    last_episode_count INTEGER,     -- 最新季的集数
    next_air_date DATE,             -- 下一集播出日期
    custom_status VARCHAR(50),      -- 自定义状态(人工纠正)
    notes TEXT,                     -- 备注
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_crawled_at TIMESTAMP
);

-- Create indexes for shows
CREATE INDEX idx_shows_tmdb_id ON shows(tmdb_id);
CREATE INDEX idx_shows_status ON shows(status);
CREATE INDEX idx_shows_last_crawled ON shows(last_crawled_at);
CREATE INDEX idx_shows_next_air_date ON shows(next_air_date);

-- ============================================
-- 2. Episodes Table (剧集详情表)
-- ============================================
CREATE TABLE IF NOT EXISTS episodes (
    id SERIAL PRIMARY KEY,
    show_id INTEGER NOT NULL REFERENCES shows(id) ON DELETE CASCADE,
    season_number INTEGER NOT NULL,
    episode_number INTEGER NOT NULL,
    name VARCHAR(255),
    overview TEXT,
    air_date DATE,
    still_path VARCHAR(512),
    runtime INTEGER,  -- in minutes
    vote_average DECIMAL(3,1),
    vote_count INTEGER,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Unique constraint
    CONSTRAINT unique_episode UNIQUE(show_id, season_number, episode_number)
);

-- Create indexes for episodes
CREATE INDEX idx_episodes_show_id ON episodes(show_id);
CREATE INDEX idx_episodes_air_date ON episodes(air_date);
CREATE INDEX idx_episodes_season ON episodes(show_id, season_number);

-- ============================================
-- 3. Crawl Logs Table (爬取日志表)
-- ============================================
CREATE TABLE IF NOT EXISTS crawl_logs (
    id SERIAL PRIMARY KEY,
    show_id INTEGER REFERENCES shows(id) ON DELETE SET NULL,
    tmdb_id INTEGER,
    
    -- Crawl information
    action VARCHAR(50),   -- 'fetch'/'refresh'/'batch'
    status VARCHAR(20),   -- 'success'/'failed'/'partial'
    episodes_count INTEGER DEFAULT 0,
    error_message TEXT,
    
    -- Performance data
    duration_ms INTEGER,
    
    -- Timestamp
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for crawl_logs
CREATE INDEX idx_crawl_logs_show_id ON crawl_logs(show_id);
CREATE INDEX idx_crawl_logs_status ON crawl_logs(status);
CREATE INDEX idx_crawl_logs_created_at ON crawl_logs(created_at);

-- ============================================
-- 4. Telegraph Posts Table (Telegraph发布记录表)
-- ============================================
CREATE TABLE IF NOT EXISTS telegraph_posts (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    telegraph_url VARCHAR(512),
    telegraph_path VARCHAR(255),   -- Telegraph article path
    content_hash VARCHAR(64),      -- Content hash (avoid duplicate publish)
    
    -- Statistics
    shows_count INTEGER DEFAULT 0,
    episodes_count INTEGER DEFAULT 0,
    date_range VARCHAR(50),        -- '2026-01-11 to 2026-02-10'
    
    -- Timestamp
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for telegraph_posts
CREATE INDEX idx_telegraph_posts_created_at ON telegraph_posts(created_at);
CREATE INDEX idx_telegraph_posts_content_hash ON telegraph_posts(content_hash);

-- ============================================
-- 5. Crawl Tasks Table (爬虫任务表)
-- ============================================
CREATE TABLE IF NOT EXISTS crawl_tasks (
    id SERIAL PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL,
    params TEXT,
    error_message TEXT,
    started_at TIMESTAMP,
    finished_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for crawl_tasks
CREATE INDEX idx_crawl_tasks_status ON crawl_tasks(status);
CREATE INDEX idx_crawl_tasks_created_at ON crawl_tasks(created_at);

-- ============================================
-- 6. Functions and Triggers
-- ============================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers to auto-update updated_at
CREATE TRIGGER update_shows_updated_at BEFORE UPDATE ON shows
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_episodes_updated_at BEFORE UPDATE ON episodes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- 6. Sample Data (Optional - for testing)
-- ============================================

-- Note: Remove this section in production

-- Insert sample show (for testing)
-- INSERT INTO shows (tmdb_id, name, original_name, status, first_air_date, overview)
-- VALUES (278577, '骄阳似我', '骄阳似我', 'Returning Series', '2025-01-01', '测试剧集简介');

-- ============================================
-- 7. Views (Optional - for easier queries)
-- ============================================

-- View: Shows with latest episode info
CREATE OR REPLACE VIEW v_shows_with_latest_episode AS
SELECT 
    s.id,
    s.tmdb_id,
    s.name,
    s.status,
    s.last_season_number,
    s.last_episode_count,
    s.next_air_date,
    s.created_at,
    s.updated_at,
    COUNT(e.id) as total_episodes
FROM shows s
LEFT JOIN episodes e ON s.id = e.show_id
GROUP BY s.id;

-- View: Recent crawl logs
CREATE OR REPLACE VIEW v_recent_crawl_logs AS
SELECT 
    cl.id,
    cl.show_id,
    s.name as show_name,
    s.tmdb_id,
    cl.action,
    cl.status,
    cl.episodes_count,
    cl.duration_ms,
    cl.created_at
FROM crawl_logs cl
LEFT JOIN shows s ON cl.show_id = s.id
ORDER BY cl.created_at DESC;

-- ============================================
-- 8. Comments
-- ============================================

COMMENT ON TABLE shows IS 'TV shows information from TMDB';
COMMENT ON TABLE episodes IS 'Episode details for each show';
COMMENT ON TABLE crawl_logs IS 'Crawling operation logs';
COMMENT ON TABLE telegraph_posts IS 'Telegraph publication records';

COMMENT ON COLUMN shows.custom_status IS 'Custom status for manual correction';
COMMENT ON COLUMN shows.notes IS 'Additional notes for the show';
COMMENT ON COLUMN episodes.runtime IS 'Episode runtime in minutes';
COMMENT ON COLUMN crawl_logs.duration_ms IS 'Crawling duration in milliseconds';
COMMENT ON COLUMN telegraph_posts.content_hash IS 'MD5 hash to avoid duplicate publishing';

-- ============================================
-- End of Schema
-- ============================================
