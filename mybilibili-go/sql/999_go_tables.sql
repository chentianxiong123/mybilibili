-- Go 代码需要但 MySQL 没有的 25 张表
-- 建在 MySQL 现有的 39 张表之后

SET session_replication_role = 'replica';

-- 1. follows
CREATE TABLE IF NOT EXISTS follows (
    id BIGSERIAL PRIMARY KEY,
    follower_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    following_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(follower_id, following_id)
);

-- 2. watch_history
CREATE TABLE IF NOT EXISTS watch_history (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    manuscript_id BIGINT NOT NULL REFERENCES manuscripts(id) ON DELETE CASCADE,
    progress_seconds INT NOT NULL DEFAULT 0,
    watched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, manuscript_id)
);

-- 3. favorite_folder_videos
CREATE TABLE IF NOT EXISTS favorite_folder_videos (
    id BIGSERIAL PRIMARY KEY,
    folder_id BIGINT NOT NULL REFERENCES favorite_folders(id) ON DELETE CASCADE,
    manuscript_id BIGINT NOT NULL REFERENCES manuscripts(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(folder_id, manuscript_id)
);

-- 4. danmaku
CREATE TABLE IF NOT EXISTS danmaku (
    id BIGSERIAL PRIMARY KEY,
    video_id BIGINT NOT NULL,
    manuscript_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    time INT NOT NULL DEFAULT 0,
    color VARCHAR(10) NOT NULL DEFAULT '#FFFFFF',
    mode INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 5. upload_sessions
CREATE TABLE IF NOT EXISTS upload_sessions (
    id VARCHAR(32) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    title VARCHAR(100) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    category_id BIGINT NOT NULL DEFAULT 0,
    tags TEXT NOT NULL DEFAULT '',
    videos JSONB NOT NULL DEFAULT '[]',
    uploaded_chunks INT NOT NULL DEFAULT 0,
    total_chunks INT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 6. live_seats
CREATE TABLE IF NOT EXISTS live_seats (
    id BIGSERIAL PRIMARY KEY,
    room_id BIGINT NOT NULL REFERENCES live_rooms(id) ON DELETE CASCADE,
    seat_index INT NOT NULL,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    status INT NOT NULL DEFAULT 0,
    muted BOOLEAN NOT NULL DEFAULT false,
    joined_at TIMESTAMPTZ,
    UNIQUE(room_id, seat_index)
);

-- 7. login_logs
CREATE TABLE IF NOT EXISTS login_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    ip VARCHAR(50) NOT NULL DEFAULT '',
    user_agent TEXT NOT NULL DEFAULT '',
    status INT NOT NULL DEFAULT 0,
    login_time TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 8. verification_codes
CREATE TABLE IF NOT EXISTS verification_codes (
    id BIGSERIAL PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL,
    code_type VARCHAR(20) NOT NULL,
    code_value VARCHAR(10) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 9. ai_api_configs (fixed: added type column)
CREATE TABLE IF NOT EXISTS ai_api_configs (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'LLM',
    base_url VARCHAR(500) NOT NULL,
    api_key VARCHAR(500) NOT NULL DEFAULT '',
    model VARCHAR(100) NOT NULL DEFAULT '',
    max_tokens INT NOT NULL DEFAULT 4096,
    temperature REAL NOT NULL DEFAULT 0.7,
    enabled BOOLEAN NOT NULL DEFAULT true,
    extra_config JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 10. ai_bindings
CREATE TABLE IF NOT EXISTS ai_bindings (
    id BIGSERIAL PRIMARY KEY,
    feature VARCHAR(50) NOT NULL UNIQUE,
    api_config_id BIGINT NOT NULL REFERENCES ai_api_configs(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 11. ai_skills
CREATE TABLE IF NOT EXISTS ai_skills (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    system_prompt TEXT NOT NULL DEFAULT '',
    few_shot_examples TEXT NOT NULL DEFAULT '',
    type VARCHAR(20) NOT NULL DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 12. ai_usage_logs
CREATE TABLE IF NOT EXISTS ai_usage_logs (
    id BIGSERIAL PRIMARY KEY,
    feature VARCHAR(50) NOT NULL,
    user_id BIGINT NOT NULL DEFAULT 0,
    token_count INT NOT NULL DEFAULT 0,
    model VARCHAR(50) NOT NULL DEFAULT '',
    duration_ms INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 13. support_tickets
CREATE TABLE IF NOT EXISTS support_tickets (
    id BIGSERIAL PRIMARY KEY,
    ticket_no VARCHAR(20) NOT NULL UNIQUE,
    user_id BIGINT NOT NULL,
    session_id BIGINT,
    source VARCHAR(20) NOT NULL DEFAULT '',
    category VARCHAR(50) NOT NULL DEFAULT '',
    priority INT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN',
    title VARCHAR(200) NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    entry_reply TEXT NOT NULL DEFAULT '',
    admin_reply TEXT NOT NULL DEFAULT '',
    assignee_admin_id BIGINT,
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 14. notifications
CREATE TABLE IF NOT EXISTS notifications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    type VARCHAR(20) NOT NULL DEFAULT '',
    title VARCHAR(200) NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    from_user_id BIGINT,
    target_id BIGINT,
    is_read INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 15. recommend_configs
CREATE TABLE IF NOT EXISTS recommend_configs (
    id BIGSERIAL PRIMARY KEY,
    config_key VARCHAR(100) NOT NULL UNIQUE,
    config_json JSONB NOT NULL DEFAULT '{}',
    updated_by VARCHAR(50) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 16. hot_search
CREATE TABLE IF NOT EXISTS hot_search (
    id BIGSERIAL PRIMARY KEY,
    keyword VARCHAR(50) NOT NULL UNIQUE,
    score INT NOT NULL DEFAULT 0,
    rank INT NOT NULL DEFAULT 0,
    search_count INT NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 17. dynamic_likes
CREATE TABLE IF NOT EXISTS dynamic_likes (
    id BIGSERIAL PRIMARY KEY,
    dynamic_id BIGINT NOT NULL REFERENCES user_dynamics(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(dynamic_id, user_id)
);

-- 18. content_reviews (fixed: added updated_at)
CREATE TABLE IF NOT EXISTS content_reviews (
    id BIGSERIAL PRIMARY KEY,
    type VARCHAR(20) NOT NULL DEFAULT '',
    user_id BIGINT NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    status INT NOT NULL DEFAULT 0,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 19. search_documents
CREATE TABLE IF NOT EXISTS search_documents (
    id TEXT NOT NULL,
    index_name TEXT NOT NULL,
    doc_json JSONB NOT NULL,
    search_vector TSVECTOR,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (index_name, id)
);

-- 20. jsonb_documents
CREATE TABLE IF NOT EXISTS jsonb_documents (
    id TEXT NOT NULL,
    collection_name TEXT NOT NULL,
    doc_json JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (collection_name, id)
);

-- 21. manuscript_status_events
CREATE TABLE IF NOT EXISTS manuscript_status_events (
    id BIGSERIAL PRIMARY KEY,
    manuscript_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    from_status INT NOT NULL DEFAULT 0,
    to_status INT NOT NULL DEFAULT 0,
    action VARCHAR(50) NOT NULL DEFAULT '',
    operator_type VARCHAR(20) NOT NULL DEFAULT '',
    operator_id BIGINT,
    reason TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 22. video_process_events
CREATE TABLE IF NOT EXISTS video_process_events (
    id BIGSERIAL PRIMARY KEY,
    video_id BIGINT NOT NULL,
    manuscript_id BIGINT NOT NULL,
    from_status INT NOT NULL DEFAULT 0,
    to_status INT NOT NULL DEFAULT 0,
    stage VARCHAR(50) NOT NULL DEFAULT '',
    progress INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 23. manuscript_edit_versions
CREATE TABLE IF NOT EXISTS manuscript_edit_versions (
    id BIGSERIAL PRIMARY KEY,
    manuscript_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    before_snapshot JSONB,
    after_snapshot JSONB,
    changed_fields JSONB,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 24. manuscript_daily_metrics
CREATE TABLE IF NOT EXISTS manuscript_daily_metrics (
    id BIGSERIAL PRIMARY KEY,
    metric_date DATE NOT NULL,
    manuscript_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    view_count INT NOT NULL DEFAULT 0,
    like_count INT NOT NULL DEFAULT 0,
    coin_count INT NOT NULL DEFAULT 0,
    collect_count INT NOT NULL DEFAULT 0,
    share_count INT NOT NULL DEFAULT 0,
    comment_count INT NOT NULL DEFAULT 0,
    danmaku_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(metric_date, manuscript_id)
);

SET session_replication_role = 'origin';