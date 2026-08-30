-- 消息表
CREATE TABLE IF NOT EXISTS messages (
    id INT PRIMARY KEY AUTO_INCREMENT,
    sender_id INT NOT NULL COMMENT '发送者ID',
    receiver_id INT NOT NULL COMMENT '接收者ID',
    content VARCHAR(500) NOT NULL COMMENT '消息内容',
    is_read TINYINT DEFAULT 0 COMMENT '是否已读：0-未读，1-已读',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (receiver_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消息表';

-- 消息阅读状态表（可选，用于更复杂的消息阅读状态管理）
CREATE TABLE IF NOT EXISTS message_read_status (
    id INT PRIMARY KEY AUTO_INCREMENT,
    message_id INT NOT NULL COMMENT '消息ID',
    user_id INT NOT NULL COMMENT '用户ID',
    is_read TINYINT DEFAULT 0 COMMENT '是否已读：0-未读，1-已读',
    read_at DATETIME COMMENT '阅读时间',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY (message_id, user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消息阅读状态表';
