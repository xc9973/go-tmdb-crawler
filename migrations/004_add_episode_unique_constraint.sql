-- Add unique constraint for episodes table (SQLite version)
-- This fixes the ON CONFLICT issue in CreateBatch operations

-- Note: SQLite doesn't support adding UNIQUE constraints to existing tables directly
-- We need to recreate the table with the constraint

-- Step 1: Create a new table with the unique constraint
CREATE TABLE IF NOT EXISTS episodes_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    show_id INTEGER NOT NULL,
    season_number INTEGER NOT NULL,
    episode_number INTEGER NOT NULL,
    name VARCHAR(255),
    overview TEXT,
    air_date DATE,
    still_path VARCHAR(512),
    runtime INTEGER,
    vote_average DECIMAL(3,1),
    vote_count INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (show_id) REFERENCES shows(id) ON DELETE CASCADE,
    UNIQUE(show_id, season_number, episode_number)
);

-- Step 2: Copy unique data from old table to new table
INSERT INTO episodes_new (
    id, show_id, season_number, episode_number, name, overview, 
    air_date, still_path, runtime, vote_average, vote_count, 
    created_at, updated_at
)
SELECT DISTINCT
    MIN(id) as id,
    show_id, season_number, episode_number, 
    MAX(name) as name,
    MAX(overview) as overview,
    MAX(air_date) as air_date,
    MAX(still_path) as still_path,
    MAX(runtime) as runtime,
    MAX(vote_average) as vote_average,
    MAX(vote_count) as vote_count,
    MAX(created_at) as created_at,
    MAX(updated_at) as updated_at
FROM episodes
GROUP BY show_id, season_number, episode_number
ORDER BY MIN(id);

-- Step 3: Drop old table
DROP TABLE episodes;

-- Step 4: Rename new table to original name
ALTER TABLE episodes_new RENAME TO episodes;

-- Step 5: Recreate indexes
CREATE INDEX IF NOT EXISTS idx_episodes_show_id ON episodes(show_id);
CREATE INDEX IF NOT EXISTS idx_episodes_air_date ON episodes(air_date);
CREATE INDEX IF NOT EXISTS idx_episodes_season ON episodes(show_id, season_number);
