-- 修复 videos 表结构，添加缺失的字段
-- 执行前请先备份数据库

-- 添加 description 字段（如果不存在）
ALTER TABLE videos ADD COLUMN IF NOT EXISTS description TEXT COMMENT '视频描述';

-- 添加 manuscript_id 字段（如果不存在）
ALTER TABLE videos ADD COLUMN IF NOT EXISTS manuscript_id INT COMMENT '所属稿件ID';

-- 添加 video_order 字段（如果不存在）
ALTER TABLE videos ADD COLUMN IF NOT EXISTS video_order INT DEFAULT 0 COMMENT '在稿件中的排序';

-- 添加 duration_seconds 字段（如果不存在）
ALTER TABLE videos ADD COLUMN IF NOT EXISTS duration_seconds INT COMMENT '时长秒数';

-- 添加 status 字段（如果不存在）
ALTER TABLE videos ADD COLUMN IF NOT EXISTS status INT DEFAULT 0 COMMENT '状态';

-- 添加 review_status 字段（如果不存在）
ALTER TABLE videos ADD COLUMN IF NOT EXISTS review_status INT DEFAULT 0 COMMENT '审核状态';

-- 添加 process_progress 字段（如果不存在）
ALTER TABLE videos ADD COLUMN IF NOT EXISTS process_progress INT DEFAULT 0 COMMENT '处理进度';

-- 添加其他可能缺失的字段
ALTER TABLE videos ADD COLUMN IF NOT EXISTS play_url_hd VARCHAR(500) COMMENT '高清播放地址';
ALTER TABLE videos ADD COLUMN IF NOT EXISTS play_url_sd VARCHAR(500) COMMENT '标清播放地址';
ALTER TABLE videos ADD COLUMN IF NOT EXISTS play_url_ld VARCHAR(500) COMMENT '流畅播放地址';
ALTER TABLE videos ADD COLUMN IF NOT EXISTS process_stage VARCHAR(50) COMMENT '处理阶段';
ALTER TABLE videos ADD COLUMN IF NOT EXISTS has_subtitle INT DEFAULT 0 COMMENT '是否有字幕';
ALTER TABLE videos ADD COLUMN IF NOT EXISTS has_summary INT DEFAULT 0 COMMENT '是否有AI总结';
ALTER TABLE videos ADD COLUMN IF NOT EXISTS review_reason VARCHAR(500) COMMENT '审核原因';
ALTER TABLE videos ADD COLUMN IF NOT EXISTS review_time DATETIME COMMENT '审核时间';
ALTER TABLE videos ADD COLUMN IF NOT EXISTS reviewer_id INT COMMENT '审核人ID';
ALTER TABLE videos ADD COLUMN IF NOT EXISTS process_status INT DEFAULT 0 COMMENT '处理状态';
ALTER TABLE videos ADD COLUMN IF NOT EXISTS process_error VARCHAR(500) COMMENT '处理错误信息';
ALTER TABLE videos ADD COLUMN IF NOT EXISTS source_video_url VARCHAR(500) COMMENT '源视频URL';
