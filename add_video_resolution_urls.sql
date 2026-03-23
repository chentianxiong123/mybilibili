-- 添加多分辨率视频URL字段
-- 执行时间: 2026-03-11
-- 说明: 为支持视频多分辨率播放，添加高清、标清、流畅三个分辨率的URL字段

-- 添加高清视频URL字段 (1080p)
ALTER TABLE videos ADD COLUMN play_url_hd VARCHAR(255) DEFAULT NULL COMMENT '高清视频URL(1080p)' AFTER play_url;

-- 添加标清视频URL字段 (720p)
ALTER TABLE videos ADD COLUMN play_url_sd VARCHAR(255) DEFAULT NULL COMMENT '标清视频URL(720p)' AFTER play_url_hd;

-- 添加流畅视频URL字段 (480p)
ALTER TABLE videos ADD COLUMN play_url_ld VARCHAR(255) DEFAULT NULL COMMENT '流畅视频URL(480p)' AFTER play_url_sd;

-- 将现有的play_url数据迁移到play_url_hd（作为默认高清）
UPDATE videos SET play_url_hd = play_url WHERE play_url IS NOT NULL AND play_url_hd IS NULL;

-- 添加索引以优化查询
-- CREATE INDEX idx_videos_status ON videos(status);
