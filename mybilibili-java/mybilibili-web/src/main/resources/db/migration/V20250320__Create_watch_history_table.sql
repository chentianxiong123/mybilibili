-- 创建浏览历史表
-- 记录用户观看视频的进度和历史

CREATE TABLE IF NOT EXISTS watch_history (
    id INT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    user_id INT NOT NULL COMMENT '用户ID',
    video_id INT NOT NULL COMMENT '视频ID',
    progress_seconds INT DEFAULT 0 COMMENT '观看进度（秒）',
    watched_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后观看时间',

    -- 联合唯一索引，防止同一用户对同一视频重复记录
    UNIQUE KEY uk_user_video (user_id, video_id),

    -- 查询索引：按用户和时间查询浏览历史
    INDEX idx_user_time (user_id, watched_at DESC)

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户浏览历史表';
