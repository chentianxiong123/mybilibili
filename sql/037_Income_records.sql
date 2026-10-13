-- 收益数据：创作者分成/充电等收入明细，数据中心「收益数据」模块读取。
-- 目前平台尚无充值/分成能力，表先行，收入为真实 0（将来接入充电打赏后写入）。
CREATE TABLE IF NOT EXISTS income_records (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    type VARCHAR(30) NOT NULL DEFAULT 'custom_charge',  -- 收入类型（如 custom_charge 自定义充电）
    amount_cents BIGINT NOT NULL DEFAULT 0,              -- 金额（分），避免浮点
    source_manuscript_id BIGINT,
    remark VARCHAR(200) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_income_records_user ON income_records(user_id, created_at);