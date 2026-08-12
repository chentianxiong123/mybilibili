ALTER TABLE live_linkmic ADD COLUMN IF NOT EXISTS viewer_name VARCHAR(50) NOT NULL DEFAULT '';
ALTER TABLE live_linkmic ADD COLUMN IF NOT EXISTS audio_enabled INT NOT NULL DEFAULT 0;
ALTER TABLE live_linkmic ADD COLUMN IF NOT EXISTS video_enabled INT NOT NULL DEFAULT 0;
ALTER TABLE live_linkmic ADD COLUMN IF NOT EXISTS max_linkmics INT NOT NULL DEFAULT 3;
ALTER TABLE live_linkmic ADD COLUMN IF NOT EXISTS apply_time TIMESTAMPTZ;
ALTER TABLE live_linkmic ADD COLUMN IF NOT EXISTS connect_time TIMESTAMPTZ;
ALTER TABLE live_linkmic ADD COLUMN IF NOT EXISTS end_time TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_live_linkmic_room ON live_linkmic(room_id, status);
CREATE INDEX IF NOT EXISTS idx_live_linkmic_viewer ON live_linkmic(viewer_id);