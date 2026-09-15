CREATE TABLE categories (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    icon TEXT NOT NULL DEFAULT '',
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE manuscripts (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    cover_url VARCHAR(500) NOT NULL DEFAULT '',
    user_id BIGINT NOT NULL REFERENCES users(id),
    category_id BIGINT NOT NULL REFERENCES categories(id),
    view_count BIGINT NOT NULL DEFAULT 0,
    like_count BIGINT NOT NULL DEFAULT 0,
    coin_count BIGINT NOT NULL DEFAULT 0,
    collect_count BIGINT NOT NULL DEFAULT 0,
    share_count BIGINT NOT NULL DEFAULT 0,
    comment_count BIGINT NOT NULL DEFAULT 0,
    danmaku_count BIGINT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 0,
    review_status INT NOT NULL DEFAULT 0,
    review_reason TEXT NOT NULL DEFAULT '',
    review_time TIMESTAMPTZ,
    reviewer_id BIGINT,
    upload_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    duration VARCHAR(20) NOT NULL DEFAULT '',
    duration_seconds INT NOT NULL DEFAULT 0
);

CREATE INDEX idx_manuscripts_user_id ON manuscripts(user_id);
CREATE INDEX idx_manuscripts_category_id ON manuscripts(category_id);
CREATE INDEX idx_manuscripts_status ON manuscripts(status);
CREATE INDEX idx_manuscripts_upload_time ON manuscripts(upload_time DESC);

CREATE TABLE videos (
    id BIGSERIAL PRIMARY KEY,
    manuscript_id BIGINT NOT NULL REFERENCES manuscripts(id) ON DELETE CASCADE,
    video_order INT NOT NULL DEFAULT 0,
    title VARCHAR(100) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    play_url_hd VARCHAR(500) NOT NULL DEFAULT '',
    play_url_sd VARCHAR(500) NOT NULL DEFAULT '',
    play_url_ld VARCHAR(500) NOT NULL DEFAULT '',
    upload_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    process_progress INT NOT NULL DEFAULT 0,
    process_stage VARCHAR(50) NOT NULL DEFAULT '',
    has_subtitle INT NOT NULL DEFAULT 0,
    has_summary INT NOT NULL DEFAULT 0,
    process_status INT NOT NULL DEFAULT 0,
    process_error VARCHAR(500) NOT NULL DEFAULT '',
    source_video_url VARCHAR(500) NOT NULL DEFAULT '',
    duration_seconds INT NOT NULL DEFAULT 0
);

CREATE INDEX idx_videos_manuscript_id ON videos(manuscript_id);
CREATE INDEX idx_videos_upload_time ON videos(upload_time DESC);

CREATE TABLE tags (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(30) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE video_tags (
    video_id BIGINT NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    tag_id BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    UNIQUE(video_id, tag_id)
);

INSERT INTO categories (id, name, icon, sort_order) VALUES
    (1, '人工智能', '', 1),
    (2, '电子', '', 2),
    (3, '数学', '', 3),
    (4, '英语', '', 4),
    (5, '运动', '', 5),
    (6, '心理学', '', 6),
    (7, '软件', '', 7),
    (8, '硬件', '', 8),
    (9, '物理', '', 9),
    (10, '机械', '', 10),
    (11, '科技', '', 11),
    (12, '政治', '', 12),
    (13, '历史', '', 13),
    (14, '经济', '', 14);
CREATE TABLE IF NOT EXISTS upload_sessions (
    id VARCHAR(32) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    title VARCHAR(100) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    category_id BIGINT,
    tags TEXT NOT NULL DEFAULT '',
    videos JSONB NOT NULL DEFAULT '[]',
    uploaded_chunks INT NOT NULL DEFAULT 0,
    total_chunks INT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
