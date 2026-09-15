CREATE TABLE IF NOT EXISTS dynamic_likes (
    id BIGSERIAL PRIMARY KEY,
    dynamic_id BIGINT NOT NULL REFERENCES user_dynamics(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(dynamic_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_dynamic_likes_dynamic ON dynamic_likes(dynamic_id);