-- 创建用户隐私设置表
CREATE TABLE IF NOT EXISTS user_privacy_settings (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL UNIQUE,
    -- 隐私设置项
    public_collection TINYINT DEFAULT 1 COMMENT '公开我的收藏: 1-公开, 0-隐藏',
    public_birthday_tags TINYINT DEFAULT 0 COMMENT '公开生日和个人标签: 1-公开, 0-隐藏',
    public_coin_videos TINYINT DEFAULT 0 COMMENT '公开最近投币的视频: 1-公开, 0-隐藏',
    public_like_videos TINYINT DEFAULT 0 COMMENT '公开最近点赞的视频: 1-公开, 0-隐藏',
    public_following_list TINYINT DEFAULT 0 COMMENT '公开我的关注列表: 1-公开, 0-隐藏',
    public_followers_list TINYINT DEFAULT 0 COMMENT '公开我的粉丝列表: 1-公开, 0-隐藏',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建用户个人标签表
CREATE TABLE IF NOT EXISTS user_tags (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    tag_name VARCHAR(20) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY (user_id, tag_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建索引
CREATE INDEX idx_user_tags_user_id ON user_tags(user_id);
