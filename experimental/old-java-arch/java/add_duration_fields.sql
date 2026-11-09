-- ============================================
-- Add duration fields to videos and manuscripts tables
-- ============================================

-- Add duration_seconds field to videos table
ALTER TABLE videos ADD COLUMN duration_seconds INT DEFAULT 0 COMMENT 'Video duration in seconds';

-- Add duration_seconds field to manuscripts table
ALTER TABLE manuscripts ADD COLUMN duration_seconds INT DEFAULT 0 COMMENT 'Total duration in seconds (sum of all videos)';

-- Initialize existing data: parse duration string to seconds
-- Convert duration format "HH:MM:SS" or "MM:SS" to seconds
UPDATE videos 
SET duration_seconds = 
    CASE 
        WHEN duration LIKE '%:%:%' THEN 
            -- Format: HH:MM:SS
            SUBSTRING_INDEX(duration, ':', 1) * 3600 + 
            SUBSTRING_INDEX(SUBSTRING_INDEX(duration, ':', 2), ':', -1) * 60 + 
            SUBSTRING_INDEX(duration, ':', -1)
        WHEN duration LIKE '%:%' THEN 
            -- Format: MM:SS
            SUBSTRING_INDEX(duration, ':', 1) * 60 + 
            SUBSTRING_INDEX(duration, ':', -1)
        ELSE 
            0
    END
WHERE duration_seconds = 0 AND duration IS NOT NULL AND duration != '';

-- Update manuscripts total duration
UPDATE manuscripts m 
SET duration_seconds = (
    SELECT COALESCE(SUM(duration_seconds), 0) 
    FROM videos v 
    WHERE v.manuscript_id = m.id
)
WHERE EXISTS (SELECT 1 FROM videos WHERE manuscript_id = m.id);
