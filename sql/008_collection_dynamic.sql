CREATE TABLE IF NOT EXISTS manuscript_collections (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    cover_url VARCHAR(500) NOT NULL DEFAULT '',
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    manuscript_count INT NOT NULL DEFAULT 0,
    view_count BIGINT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_collections_user ON manuscript_collections(user_id);
CREATE INDEX IF NOT EXISTS idx_collections_status ON manuscript_collections(status);

CREATE TABLE IF NOT EXISTS manuscript_collection_relations (
    id BIGSERIAL PRIMARY KEY,
    manuscript_id BIGINT NOT NULL REFERENCES manuscripts(id) ON DELETE CASCADE,
    collection_id BIGINT NOT NULL REFERENCES manuscript_collections(id) ON DELETE CASCADE,
    collection_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(manuscript_id, collection_id)
);

CREATE INDEX IF NOT EXISTS idx_collection_relations_coll ON manuscript_collection_relations(collection_id, collection_order);

CREATE TABLE IF NOT EXISTS user_dynamics (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    dynamic_type INT NOT NULL DEFAULT 0,
    image_url VARCHAR(2000) NOT NULL DEFAULT '',
    ref_manuscript_id BIGINT,
    like_count INT NOT NULL DEFAULT 0,
    comment_count INT NOT NULL DEFAULT 0,
    share_count INT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_dynamics_user ON user_dynamics(user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS dynamic_comments (
    id BIGSERIAL PRIMARY KEY,
    dynamic_id BIGINT NOT NULL REFERENCES user_dynamics(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    parent_id BIGINT,
    reply_user_id BIGINT,
    like_count INT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_dynamic_comments_dyn ON dynamic_comments(dynamic_id);
CREATE INDEX IF NOT EXISTS idx_dynamic_comments_parent ON dynamic_comments(parent_id);

CREATE TABLE IF NOT EXISTS shares (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT,
    manuscript_id BIGINT NOT NULL REFERENCES manuscripts(id) ON DELETE CASCADE,
    channel VARCHAR(50) NOT NULL DEFAULT 'unknown',
    ip_address VARCHAR(50) NOT NULL DEFAULT '',
    create_time TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shares_manuscript ON shares(manuscript_id);