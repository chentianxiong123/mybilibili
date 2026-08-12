CREATE TABLE IF NOT EXISTS prohibited_words (
    id BIGSERIAL PRIMARY KEY,
    word VARCHAR(255) NOT NULL,
    match_type VARCHAR(20) NOT NULL DEFAULT 'CONTAINS',
    category VARCHAR(50) NOT NULL DEFAULT '',
    is_enabled INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_prohibited_words_word ON prohibited_words(word);
CREATE INDEX IF NOT EXISTS idx_prohibited_words_category ON prohibited_words(category);

CREATE TABLE IF NOT EXISTS reports (
    id BIGSERIAL PRIMARY KEY,
    reporter_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_type VARCHAR(30) NOT NULL,
    target_id BIGINT NOT NULL,
    manuscript_id BIGINT,
    reason VARCHAR(50) NOT NULL,
    description VARCHAR(500) NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    admin_remark VARCHAR(500) NOT NULL DEFAULT '',
    processed_at TIMESTAMPTZ,
    ai_review_status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    ai_verdict VARCHAR(500) NOT NULL DEFAULT '',
    ai_risk_level VARCHAR(10) NOT NULL DEFAULT '',
    ai_reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_reports_status ON reports(status);
CREATE INDEX IF NOT EXISTS idx_reports_target ON reports(target_type, target_id);

CREATE TABLE IF NOT EXISTS banner_images (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL DEFAULT '',
    image_url VARCHAR(500) NOT NULL DEFAULT '',
    link_url VARCHAR(500) NOT NULL DEFAULT '',
    sort_order INT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 1,
    type INT NOT NULL,
    category_id BIGINT,
    start_time TIMESTAMPTZ,
    end_time TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_banners_type ON banner_images(type);
CREATE INDEX IF NOT EXISTS idx_banners_status ON banner_images(status);