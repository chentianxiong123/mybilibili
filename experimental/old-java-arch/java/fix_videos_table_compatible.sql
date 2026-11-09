-- 修复 videos 表结构，添加缺失的字段
-- 兼容所有 MySQL 版本（包括不支持 IF NOT EXISTS 的版本）
-- 使用方法：在 IDEA Database 控制台或 MySQL 客户端中执行

-- 先选择数据库
USE mybilibili;

-- ============================================
-- 第1步：添加 insert 语句必需的字段
-- ============================================

-- 添加 description 字段
ALTER TABLE videos ADD COLUMN description TEXT COMMENT '视频描述';

-- 添加 manuscript_id 字段
ALTER TABLE videos ADD COLUMN manuscript_id INT COMMENT '所属稿件ID';

-- 添加 video_order 字段
ALTER TABLE videos ADD COLUMN video_order INT DEFAULT 0 COMMENT '在稿件中的排序';

-- 添加 duration_seconds 字段
ALTER TABLE videos ADD COLUMN duration_seconds INT COMMENT '时长秒数';

-- 添加 status 字段
ALTER TABLE videos ADD COLUMN status INT DEFAULT 0 COMMENT '状态';

-- 添加 review_status 字段
ALTER TABLE videos ADD COLUMN review_status INT DEFAULT 0 COMMENT '审核状态';

-- 添加 process_progress 字段
ALTER TABLE videos ADD COLUMN process_progress INT DEFAULT 0 COMMENT '处理进度';

-- ============================================
-- 第2步：添加 update 语句使用的字段
-- ============================================

-- 添加多清晰度播放地址字段
ALTER TABLE videos ADD COLUMN play_url_hd VARCHAR(500) COMMENT '高清播放地址';
ALTER TABLE videos ADD COLUMN play_url_sd VARCHAR(500) COMMENT '标清播放地址';
ALTER TABLE videos ADD COLUMN play_url_ld VARCHAR(500) COMMENT '流畅播放地址';

-- 添加处理相关字段
ALTER TABLE videos ADD COLUMN process_stage VARCHAR(50) COMMENT '处理阶段';
ALTER TABLE videos ADD COLUMN has_subtitle INT DEFAULT 0 COMMENT '是否有字幕';
ALTER TABLE videos ADD COLUMN has_summary INT DEFAULT 0 COMMENT '是否有AI总结';

-- 添加审核相关字段
ALTER TABLE videos ADD COLUMN review_reason VARCHAR(500) COMMENT '审核原因';
ALTER TABLE videos ADD COLUMN review_time DATETIME COMMENT '审核时间';
ALTER TABLE videos ADD COLUMN reviewer_id INT COMMENT '审核人ID';

-- 添加新处理状态字段
ALTER TABLE videos ADD COLUMN process_status INT DEFAULT 0 COMMENT '处理状态：0待处理 1转码中 2音频提取中 3字幕生成中 4AI总结中 5完成';
ALTER TABLE videos ADD COLUMN process_error VARCHAR(500) COMMENT '处理错误信息';
ALTER TABLE videos ADD COLUMN source_video_url VARCHAR(500) COMMENT '源视频URL';

-- ============================================
-- 第3步：验证修复结果
-- ============================================
DESCRIBE videos;
