-- 创建通知表
CREATE TABLE IF NOT EXISTS notifications (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    type INT NOT NULL COMMENT '通知类型：1-互动通知，2-系统通知，3-私信通知，4-视频通知',
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    related_id VARCHAR(100) COMMENT '相关资源ID，如视频ID、评论ID等',
    is_read INT DEFAULT 0 COMMENT '0-未读，1-已读',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    read_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 创建通知设置表
CREATE TABLE IF NOT EXISTS notification_settings (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL UNIQUE,
    like_notification INT DEFAULT 1 COMMENT '点赞通知：1-开启，0-关闭',
    comment_notification INT DEFAULT 1 COMMENT '评论通知：1-开启，0-关闭',
    follow_notification INT DEFAULT 1 COMMENT '关注通知：1-开启，0-关闭',
    system_notification INT DEFAULT 1 COMMENT '系统通知：1-开启，0-关闭',
    video_notification INT DEFAULT 1 COMMENT '视频通知：1-开启，0-关闭',
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 创建索引
CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_is_read ON notifications(is_read);
CREATE INDEX idx_notifications_created_at ON notifications(created_at);
