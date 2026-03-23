-- 创建稿件合集表
CREATE TABLE IF NOT EXISTS manuscript_collections (
    id INT PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(100) NOT NULL COMMENT '合集标题',
    description TEXT COMMENT '合集描述',
    cover_url VARCHAR(500) COMMENT '封面图片URL',
    user_id INT NOT NULL COMMENT '创建用户ID',
    manuscript_count INT DEFAULT 0 COMMENT '稿件数量',
    view_count INT DEFAULT 0 COMMENT '浏览次数',
    status TINYINT DEFAULT 1 COMMENT '状态：0-私密，1-公开',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='稿件合集表';
