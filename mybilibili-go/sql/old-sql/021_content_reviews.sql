CREATE TABLE IF NOT EXISTS content_reviews (
    id BIGSERIAL PRIMARY KEY,
    type VARCHAR(20) NOT NULL,
    user_id BIGINT,
    content TEXT NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_content_reviews_type ON content_reviews(type);