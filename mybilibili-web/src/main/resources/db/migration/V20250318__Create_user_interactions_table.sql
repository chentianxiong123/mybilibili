-- 数据库重构：创建统一交互记录表
-- 合并 likes, dynamic_likes, dynamic_comment_likes, collections, follows, coins 表

CREATE TABLE IF NOT EXISTS user_interactions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL COMMENT '用户ID',
    target_type VARCHAR(20) NOT NULL COMMENT '目标类型：VIDEO/DYNAMIC/COMMENT/REPLY/USER',
    target_id INT NOT NULL COMMENT '目标ID',
    interaction_type VARCHAR(20) NOT NULL COMMENT '交互类型：LIKE/COLLECT/FOLLOW/COIN',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    
    -- 联合唯一索引，防止重复记录
    UNIQUE KEY uk_user_interaction (user_id, target_type, target_id, interaction_type),
    
    -- 查询索引：按目标查询交互记录
    INDEX idx_target (target_type, target_id, interaction_type),
    
    -- 查询索引：按用户查询交互记录
    INDEX idx_user (user_id, interaction_type, created_at),
    
    -- 复合索引：支持批量查询
    INDEX idx_user_target (user_id, target_type, target_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户交互记录表';

-- 创建交互计数表（用于性能优化，可选）
CREATE TABLE IF NOT EXISTS interaction_counts (
    target_type VARCHAR(20) NOT NULL COMMENT '目标类型',
    target_id INT NOT NULL COMMENT '目标ID',
    like_count INT DEFAULT 0 COMMENT '点赞数',
    collect_count INT DEFAULT 0 COMMENT '收藏数',
    share_count INT DEFAULT 0 COMMENT '分享数',
    coin_count INT DEFAULT 0 COMMENT '投币数',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    PRIMARY KEY (target_type, target_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='交互计数表';
