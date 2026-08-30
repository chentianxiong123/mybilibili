-- ============================================
-- Video processing status field migration
-- Use status code instead of percentage progress
-- Process data (current_video_id, process_queue) stored in Redis
-- ============================================

-- Add fields to videos table
ALTER TABLE videos ADD COLUMN process_status INT DEFAULT 0 COMMENT 
    'Processing status: 0-pending 1-transcoding 2-audio extracting 3-subtitle generating 4-AI summarizing 5-completed 6-transcode failed 7-audio failed 8-subtitle failed 9-AI failed';

ALTER TABLE videos ADD COLUMN process_error VARCHAR(500) COMMENT 'Processing failure reason';

ALTER TABLE videos ADD COLUMN source_video_url VARCHAR(500) COMMENT 'Source video URL (for admin preview)';

-- Note: current_video_id and process_queue are stored in Redis, not in MySQL
-- Redis Key format:
--   manuscript:process:{manuscriptId}:current  - Currently processing video ID
--   manuscript:process:{manuscriptId}:queue    - Pending video ID queue

-- Initialize existing data
-- Convert existing process_progress and process_stage to process_status
UPDATE videos SET process_status = 
    CASE 
        WHEN status = 0 THEN 0  -- Pending review
        WHEN status = 1 THEN 1  -- Processing, default to transcoding
        WHEN status = 2 THEN 5  -- Ready to publish, processing completed
        WHEN status = 3 THEN 5  -- Published, processing completed
        WHEN status = 5 THEN 6  -- Processing failed, default to transcode failed
        ELSE 0
    END
WHERE process_status IS NULL OR process_status = 0;

-- Update more precise status based on process_stage
UPDATE videos SET process_status = 1 WHERE process_stage = 'VIDEO_TRANSCODE' AND status = 1;
UPDATE videos SET process_status = 2 WHERE process_stage = 'AUDIO_EXTRACT' AND status = 1;
UPDATE videos SET process_status = 3 WHERE process_stage = 'SUBTITLE_GENERATE' AND status = 1;
UPDATE videos SET process_status = 4 WHERE process_stage = 'AI_SUMMARY' AND status = 1;
UPDATE videos SET process_status = 5 WHERE process_stage = 'COMPLETED' AND status = 2;

-- Copy play_url to source_video_url (if empty)
UPDATE videos SET source_video_url = play_url WHERE source_video_url IS NULL AND play_url IS NOT NULL;
