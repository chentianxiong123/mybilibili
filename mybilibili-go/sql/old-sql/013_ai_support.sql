CREATE TABLE IF NOT EXISTS ai_api_configs (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    base_url VARCHAR(255) NOT NULL,
    api_key VARCHAR(255) NOT NULL DEFAULT '',
    model VARCHAR(100) NOT NULL DEFAULT 'deepseek-chat',
    max_tokens INT NOT NULL DEFAULT 2000,
    temperature DOUBLE PRECISION NOT NULL DEFAULT 0.7,
    enabled INT NOT NULL DEFAULT 1,
    extra_config TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ai_bindings (
    id BIGSERIAL PRIMARY KEY,
    feature VARCHAR(30) UNIQUE NOT NULL,
    api_config_id BIGINT NOT NULL REFERENCES ai_api_configs(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS ai_skills (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500) NOT NULL DEFAULT '',
    system_prompt TEXT NOT NULL DEFAULT '',
    few_shot_examples TEXT NOT NULL DEFAULT '',
    type VARCHAR(40) NOT NULL DEFAULT '',
    enabled INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ai_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    skill_id BIGINT,
    type VARCHAR(40) NOT NULL DEFAULT 'CUSTOMER_SERVICE',
    status INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ai_sessions_user ON ai_sessions(user_id, type);

CREATE TABLE IF NOT EXISTS ai_chat_messages (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT REFERENCES ai_sessions(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL,
    content TEXT NOT NULL,
    token_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ai_messages_session ON ai_chat_messages(session_id);

CREATE TABLE IF NOT EXISTS support_tickets (
    id BIGSERIAL PRIMARY KEY,
    ticket_no VARCHAR(40) UNIQUE NOT NULL,
    user_id BIGINT,
    session_id BIGINT,
    source VARCHAR(40) NOT NULL DEFAULT 'USER_FEEDBACK',
    category VARCHAR(40) NOT NULL DEFAULT 'GENERAL',
    priority VARCHAR(20) NOT NULL DEFAULT 'NORMAL',
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    entry_reply TEXT NOT NULL DEFAULT '',
    admin_reply TEXT NOT NULL DEFAULT '',
    assignee_admin_id BIGINT,
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tickets_user ON support_tickets(user_id);
CREATE INDEX IF NOT EXISTS idx_tickets_status ON support_tickets(status);

CREATE TABLE IF NOT EXISTS ai_usage_logs (
    id BIGSERIAL PRIMARY KEY,
    feature VARCHAR(30) NOT NULL,
    user_id BIGINT,
    token_count INT NOT NULL DEFAULT 0,
    model VARCHAR(100) NOT NULL DEFAULT '',
    duration_ms INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_usage_logs_feature ON ai_usage_logs(feature, created_at);

CREATE TABLE IF NOT EXISTS notifications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    type VARCHAR(30) NOT NULL,
    title VARCHAR(200) NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    from_user_id BIGINT,
    target_id BIGINT,
    is_read INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications(user_id, is_read, created_at DESC);