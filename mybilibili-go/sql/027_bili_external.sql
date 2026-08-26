-- B站外部视频源支持（混合内容方案）
-- source_type: local(本地转码) | bilibili(B站代理)
ALTER TABLE manuscripts
    ADD COLUMN IF NOT EXISTS source_type VARCHAR(20) NOT NULL DEFAULT 'local',
    ADD COLUMN IF NOT EXISTS bvid VARCHAR(20) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS origin_url VARCHAR(500) NOT NULL DEFAULT '';

ALTER TABLE videos
    ADD COLUMN IF NOT EXISTS cid BIGINT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_manuscripts_source_type ON manuscripts(source_type);
CREATE UNIQUE INDEX IF NOT EXISTS idx_manuscripts_bvid ON manuscripts(bvid) WHERE bvid <> '';