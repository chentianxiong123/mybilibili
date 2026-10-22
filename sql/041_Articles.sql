-- 专栏（图文）投稿：草稿/发布/阅读
-- status: 0=草稿 1=已发布 -1=已删除
-- visibility: 1=所有人可见 0=仅自己可见

CREATE TABLE IF NOT EXISTS articles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(100) NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    summary VARCHAR(500) NOT NULL DEFAULT '',
    cover_url VARCHAR(500) NOT NULL DEFAULT '',
    topics VARCHAR(500) NOT NULL DEFAULT '',
    collection_name VARCHAR(100) NOT NULL DEFAULT '',
    status INT NOT NULL DEFAULT 0,
    visibility INT NOT NULL DEFAULT 1,
    comment_enabled INT NOT NULL DEFAULT 1,
    featured_comment INT NOT NULL DEFAULT 0,
    is_original INT NOT NULL DEFAULT 0,
    ai_declaration INT NOT NULL DEFAULT 0,
    word_count INT NOT NULL DEFAULT 0,
    view_count BIGINT NOT NULL DEFAULT 0,
    like_count BIGINT NOT NULL DEFAULT 0,
    favorite_count BIGINT NOT NULL DEFAULT 0,
    comment_count BIGINT NOT NULL DEFAULT 0,
    coin_count BIGINT NOT NULL DEFAULT 0,
    scheduled_at TIMESTAMPTZ,
    publish_time TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_articles_user ON articles(user_id, status, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_articles_feed ON articles(status, publish_time DESC);
