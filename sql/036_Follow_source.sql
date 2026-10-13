-- 关注来源：记录因哪个稿件产生的关注，供数据中心「涨粉量」归因使用。
-- 存量数据来源未知，保持 NULL。
ALTER TABLE follows ADD COLUMN IF NOT EXISTS source_manuscript_id BIGINT;
CREATE INDEX IF NOT EXISTS idx_follows_source ON follows(source_manuscript_id) WHERE source_manuscript_id IS NOT NULL;
