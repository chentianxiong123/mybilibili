-- 025_partition_log_tables.sql
-- 为业务日志表添加时间分区（RANGE BY month）
-- 目标：防止单表无限增长，方便按月清理，适合老硬件
-- 注意：当前表已清空 (TRUNCATE)，直接重建为分区表安全。
-- 生产有数据时需用 pg_dump + 转换或 pg_partman 扩展。

-- ========== login_logs ==========
DROP TABLE IF EXISTS login_logs CASCADE;

CREATE TABLE login_logs (
    id BIGSERIAL,
    user_id BIGINT NOT NULL,
    ip VARCHAR(50) NOT NULL DEFAULT '',
    user_agent TEXT NOT NULL DEFAULT '',
    status INT NOT NULL DEFAULT 0,
    login_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, login_time)
) PARTITION BY RANGE (login_time);

-- 示例分区（按需扩展，建议用脚本每月预创建）
CREATE TABLE IF NOT EXISTS login_logs_y2026m08 PARTITION OF login_logs
    FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');
CREATE TABLE IF NOT EXISTS login_logs_y2026m09 PARTITION OF login_logs
    FOR VALUES FROM ('2026-09-01') TO ('2026-10-01');
CREATE TABLE IF NOT EXISTS login_logs_y2026m10 PARTITION OF login_logs
    FOR VALUES FROM ('2026-10-01') TO ('2026-11-01');

CREATE INDEX IF NOT EXISTS idx_login_logs_user ON login_logs(user_id, login_time DESC);

-- ========== audit_logs ==========
DROP TABLE IF EXISTS audit_logs CASCADE;

CREATE TABLE audit_logs (
    id BIGSERIAL,
    operator_id BIGINT,
    operator_name VARCHAR(64) NOT NULL DEFAULT '',
    operator_role VARCHAR(64) NOT NULL DEFAULT '',
    module VARCHAR(64) NOT NULL,
    action VARCHAR(64) NOT NULL,
    target_type VARCHAR(64) NOT NULL DEFAULT '',
    target_id VARCHAR(128) NOT NULL DEFAULT '',
    request_method VARCHAR(16) NOT NULL DEFAULT '',
    request_uri VARCHAR(255) NOT NULL DEFAULT '',
    client_ip VARCHAR(64) NOT NULL DEFAULT '',
    user_agent VARCHAR(512) NOT NULL DEFAULT '',
    result INT NOT NULL DEFAULT 1,
    message VARCHAR(255) NOT NULL DEFAULT '',
    detail TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

CREATE TABLE IF NOT EXISTS audit_logs_y2026m08 PARTITION OF audit_logs
    FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');
CREATE TABLE IF NOT EXISTS audit_logs_y2026m09 PARTITION OF audit_logs
    FOR VALUES FROM ('2026-09-01') TO ('2026-10-01');

CREATE INDEX IF NOT EXISTS idx_audit_logs_operator ON audit_logs(operator_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_module ON audit_logs(module, action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created ON audit_logs(created_at DESC);

-- ========== ai_usage_logs ==========
DROP TABLE IF EXISTS ai_usage_logs CASCADE;

CREATE TABLE ai_usage_logs (
    id BIGSERIAL,
    feature VARCHAR(50) NOT NULL,
    user_id BIGINT NOT NULL DEFAULT 0,
    token_count INT NOT NULL DEFAULT 0,
    model VARCHAR(50) NOT NULL DEFAULT '',
    duration_ms INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

CREATE TABLE IF NOT EXISTS ai_usage_logs_y2026m08 PARTITION OF ai_usage_logs
    FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');
CREATE TABLE IF NOT EXISTS ai_usage_logs_y2026m09 PARTITION OF ai_usage_logs
    FOR VALUES FROM ('2026-09-01') TO ('2026-10-01');

CREATE INDEX IF NOT EXISTS idx_usage_logs_feature ON ai_usage_logs(feature, created_at);

-- 保留策略建议（可在 scheduler 或 cron 里执行）：
-- 保留 login_logs 最近 6 个月，audit_logs 12 个月，ai_usage_logs 3 个月。
-- 示例（需手动/定时）：
--   DROP TABLE IF EXISTS login_logs_y2026m01;
--   ... 旧分区

COMMENT ON TABLE login_logs IS '用户登录历史（分区表，按 login_time 月度）';
COMMENT ON TABLE audit_logs IS '管理员审计日志（分区表，按 created_at 月度）';
COMMENT ON TABLE ai_usage_logs IS 'AI 用量日志（分区表，按 created_at 月度）';
