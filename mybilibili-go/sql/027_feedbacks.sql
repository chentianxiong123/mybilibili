-- 用户反馈表（区别于举报 reports：无目标对象，纯用户反馈）
CREATE TABLE IF NOT EXISTS feedbacks (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL DEFAULT 0,
    type VARCHAR(30) NOT NULL DEFAULT 'OTHER',
    content TEXT NOT NULL,
    contact VARCHAR(100) NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_feedbacks_created_at ON feedbacks(created_at);
CREATE INDEX IF NOT EXISTS idx_feedbacks_status ON feedbacks(status);