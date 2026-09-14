CREATE TABLE IF NOT EXISTS manuscript_edit_versions (
    id BIGSERIAL PRIMARY KEY,
    manuscript_id BIGINT NOT NULL REFERENCES manuscripts(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    before_snapshot TEXT NOT NULL DEFAULT '',
    after_snapshot TEXT NOT NULL DEFAULT '',
    changed_fields TEXT NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
    reviewer_id BIGINT,
    review_reason VARCHAR(500) NOT NULL DEFAULT '',
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_edit_versions_ms ON manuscript_edit_versions(manuscript_id);
CREATE INDEX IF NOT EXISTS idx_edit_versions_status ON manuscript_edit_versions(status);

CREATE TABLE IF NOT EXISTS manuscript_status_events (
    id BIGSERIAL PRIMARY KEY,
    manuscript_id BIGINT NOT NULL REFERENCES manuscripts(id) ON DELETE CASCADE,
    user_id BIGINT,
    from_status INT,
    to_status INT NOT NULL,
    action VARCHAR(64) NOT NULL,
    operator_type VARCHAR(32) NOT NULL DEFAULT 'SYSTEM',
    operator_id BIGINT,
    reason VARCHAR(500) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_status_events_ms ON manuscript_status_events(manuscript_id, created_at);

CREATE TABLE IF NOT EXISTS manuscript_daily_metrics (
    metric_date DATE NOT NULL,
    manuscript_id BIGINT NOT NULL REFERENCES manuscripts(id) ON DELETE CASCADE,
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
    PRIMARY KEY(metric_date, manuscript_id)
);

CREATE INDEX IF NOT EXISTS idx_daily_metrics_user ON manuscript_daily_metrics(user_id, metric_date);

CREATE TABLE IF NOT EXISTS video_process_events (
    id BIGSERIAL PRIMARY KEY,
    video_id BIGINT NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    manuscript_id BIGINT,
    from_status INT,
    to_status INT NOT NULL,
    stage VARCHAR(64) NOT NULL DEFAULT '',
    progress INT NOT NULL DEFAULT 0,
    error_message VARCHAR(500) NOT NULL DEFAULT '',
    operator_type VARCHAR(32) NOT NULL DEFAULT 'SYSTEM',
    operator_id BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_process_events_video ON video_process_events(video_id, created_at);