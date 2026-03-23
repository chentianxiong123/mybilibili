-- ============================================
-- 创作者设置表
-- ============================================

CREATE TABLE IF NOT EXISTS creator_settings (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL UNIQUE COMMENT '用户ID',
    default_category_id INT COMMENT '默认投稿分类ID',
    auto_publish TINYINT DEFAULT 0 COMMENT '自动发布：1-开启，0-关闭',
    comment_notify TINYINT DEFAULT 1 COMMENT '评论通知：1-开启，0-关闭',
    like_notify TINYINT DEFAULT 1 COMMENT '点赞通知：1-开启，0-关闭',
    follow_notify TINYINT DEFAULT 1 COMMENT '关注通知：1-开启，0-关闭',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (default_category_id) REFERENCES categories(id) ON DELETE SET NULL,
    INDEX idx_user_id (user_id),
    INDEX idx_default_category_id (default_category_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='创作者设置表';
