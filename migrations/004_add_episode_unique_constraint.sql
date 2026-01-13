-- Add unique constraint for episodes table
-- This fixes the ON CONFLICT issue in CreateBatch operations

-- First, remove any duplicate episodes (keep the one with lowest id)
WITH ranked_episodes AS (
    SELECT 
        id,
        show_id,
        season_number,
        episode_number,
        ROW_NUMBER() OVER (
            PARTITION BY show_id, season_number, episode_number 
            ORDER BY id ASC
        ) as rn
    FROM episodes
)
DELETE FROM episodes
WHERE id IN (
    SELECT id FROM ranked_episodes WHERE rn > 1
);

-- Add the unique constraint
ALTER TABLE episodes 
ADD CONSTRAINT unique_episode UNIQUE(show_id, season_number, episode_number);
