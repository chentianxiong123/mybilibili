CREATE TABLE IF NOT EXISTS hot_search (
    id BIGSERIAL PRIMARY KEY,
    keyword VARCHAR(100) NOT NULL UNIQUE,
    score INT NOT NULL DEFAULT 0,
    rank INT NOT NULL DEFAULT 0,
    search_count INT NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_hot_search_score ON hot_search(score DESC);
CREATE INDEX IF NOT EXISTS idx_hot_search_rank ON hot_search(rank ASC);
CREATE INDEX IF NOT EXISTS idx_hot_search_expires ON hot_search(expires_at);