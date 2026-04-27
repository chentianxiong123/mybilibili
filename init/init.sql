CREATE DATABASE IF NOT EXISTS mybilibili_cloud DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE mybilibili_cloud;

CREATE TABLE IF NOT EXISTS users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    nickname VARCHAR(50) NOT NULL,
    avatar VARCHAR(255),
    email VARCHAR(100) UNIQUE,
    phone VARCHAR(20) UNIQUE,
    gender TINYINT DEFAULT 0,
    birthdate DATE,
    level INT DEFAULT 1,
    following_count INT DEFAULT 0,
    follower_count INT DEFAULT 0,
    video_count INT DEFAULT 0,
    liked_count INT DEFAULT 0,
    coin_count INT DEFAULT 0,
    experience INT DEFAULT 0,
    bio TEXT,
    announcement TEXT,
    status INT DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS manuscripts (
    id INT PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(100) NOT NULL,
    description TEXT,
    cover_url VARCHAR(500),
    user_id INT NOT NULL,
    category_id INT NOT NULL,
    view_count INT DEFAULT 0,
    like_count INT DEFAULT 0,
    coin_count INT DEFAULT 0,
    collect_count INT DEFAULT 0,
    share_count INT DEFAULT 0,
    comment_count INT DEFAULT 0,
    danmaku_count INT DEFAULT 0,
    status INT DEFAULT 0,
    review_status INT DEFAULT 0,
    review_reason VARCHAR(500),
    review_time DATETIME,
    reviewer_id INT,
    upload_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS videos (
    id INT PRIMARY KEY AUTO_INCREMENT,
    manuscript_id INT,
    video_order INT DEFAULT 0,
    title VARCHAR(100) NOT NULL,
    play_url_hd VARCHAR(255),
    play_url_sd VARCHAR(255),
    play_url_ld VARCHAR(255),
    source_video_url VARCHAR(500),
    duration_seconds INT DEFAULT 0,
    description TEXT,
    cover_url VARCHAR(255),
    has_subtitle TINYINT DEFAULT 0,
    has_summary TINYINT DEFAULT 0,
    process_status INT DEFAULT 0,
    process_progress INT DEFAULT 0,
    process_stage VARCHAR(50),
    process_error VARCHAR(500),
    upload_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS comments (
    id INT PRIMARY KEY AUTO_INCREMENT,
    manuscript_id INT,
    user_id INT NOT NULL,
    target_type VARCHAR(20) DEFAULT 'VIDEO',
    target_id INT DEFAULT 0,
    content TEXT NOT NULL,
    like_count INT DEFAULT 0,
    reply_count INT DEFAULT 0,
    status INT DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS user_interactions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    target_type VARCHAR(20) NOT NULL,
    target_id INT NOT NULL,
    interaction_type VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_type (user_id, target_type),
    INDEX idx_target (target_type, target_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS messages (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    sender_id INT NOT NULL,
    receiver_id INT NOT NULL,
    message_type VARCHAR(20) DEFAULT 'TEXT',
    target_id INT,
    content TEXT NOT NULL,
    media_url VARCHAR(500),
    conversation_id BIGINT,
    is_read INT DEFAULT 0,
    status INT DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_sender (sender_id),
    INDEX idx_receiver (receiver_id),
    INDEX idx_conversation (conversation_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS conversations (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    target_user_id INT NOT NULL,
    last_message_id BIGINT,
    last_message_content TEXT,
    last_message_time DATETIME,
    unread_count INT DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_target_user (target_user_id),
    UNIQUE KEY uk_user_target (user_id, target_user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS message_settings (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL UNIQUE,
    private_message_notify TINYINT(1) DEFAULT 1,
    reply_notify TINYINT(1) DEFAULT 1,
    at_notify TINYINT(1) DEFAULT 1,
    like_notify TINYINT(1) DEFAULT 1,
    system_notify TINYINT(1) DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS notifications (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    notification_type VARCHAR(20) NOT NULL,
    title VARCHAR(100) NOT NULL,
    content TEXT,
    is_read INT DEFAULT 0,
    status INT DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS coins (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    manuscript_id INT NOT NULL,
    coin_count INT DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_manuscript_id (manuscript_id),
    UNIQUE KEY uk_user_manuscript (user_id, manuscript_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS shares (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT,
    manuscript_id INT NOT NULL,
    channel VARCHAR(50) DEFAULT 'unknown',
    ip_address VARCHAR(50),
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_manuscript_id (manuscript_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS favorite_folders (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    name VARCHAR(100) NOT NULL,
    video_count INT DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS favorite_videos (
    id INT PRIMARY KEY AUTO_INCREMENT,
    folder_id INT NOT NULL,
    manuscript_id INT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_folder_id (folder_id),
    INDEX idx_manuscript_id (manuscript_id),
    UNIQUE KEY uk_folder_manuscript (folder_id, manuscript_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS prohibited_words (
    id INT PRIMARY KEY AUTO_INCREMENT,
    word VARCHAR(255) NOT NULL,
    match_type VARCHAR(20) DEFAULT 'CONTAINS',
    category VARCHAR(50),
    is_enabled INT DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_word (word),
    INDEX idx_category (category),
    INDEX idx_is_enabled (is_enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
