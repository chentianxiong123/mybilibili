CREATE TABLE comments (
    id BIGSERIAL PRIMARY KEY,
    manuscript_id BIGINT NOT NULL REFERENCES manuscripts(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    like_count INT NOT NULL DEFAULT 0,
    reply_count INT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_comments_manuscript_id ON comments(manuscript_id);
CREATE INDEX idx_comments_user_id ON comments(user_id);

CREATE TABLE replies (
    id BIGSERIAL PRIMARY KEY,
    comment_id BIGINT NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reply_to_user_id BIGINT REFERENCES users(id),
    content TEXT NOT NULL,
    like_count INT NOT NULL DEFAULT 0,
    status VARCHAR(10) NOT NULL DEFAULT 'NORMAL',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_replies_comment_id ON replies(comment_id);
CREATE INDEX idx_replies_user_id ON replies(user_id);

CREATE TABLE user_interactions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_type VARCHAR(20) NOT NULL,
    target_id BIGINT NOT NULL,
    interaction_type VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, target_type, target_id, interaction_type)
);

CREATE INDEX idx_user_interactions_target ON user_interactions(target_type, target_id, interaction_type);
CREATE INDEX idx_user_interactions_user ON user_interactions(user_id, interaction_type, created_at);