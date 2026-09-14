-- Scheduled Tasks for admin management
ALTER TABLE IF EXISTS scheduled_tasks DROP COLUMN IF EXISTS task_type;
ALTER TABLE IF EXISTS scheduled_tasks DROP COLUMN IF EXISTS task_config;
ALTER TABLE IF EXISTS scheduled_tasks DROP COLUMN IF EXISTS last_run_at;
ALTER TABLE IF EXISTS scheduled_tasks DROP COLUMN IF EXISTS last_run_result;
ALTER TABLE IF EXISTS scheduled_tasks DROP COLUMN IF EXISTS next_run_at;
ALTER TABLE IF EXISTS scheduled_tasks DROP COLUMN IF EXISTS run_count;
ALTER TABLE IF EXISTS scheduled_tasks DROP COLUMN IF EXISTS max_retries;
ALTER TABLE IF EXISTS scheduled_tasks DROP COLUMN IF EXISTS retry_count;
ALTER TABLE IF EXISTS scheduled_tasks DROP COLUMN IF EXISTS timeout_seconds;

CREATE TABLE IF NOT EXISTS scheduled_tasks (
    id BIGSERIAL PRIMARY KEY,
    task_key VARCHAR(100) NOT NULL UNIQUE,
    task_name VARCHAR(200) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    cron_expr VARCHAR(100) NOT NULL DEFAULT '',
    task_type VARCHAR(50) NOT NULL DEFAULT '',
    task_config TEXT NOT NULL DEFAULT '',
    enabled INT NOT NULL DEFAULT 1,
    last_run_at TIMESTAMPTZ,
    last_run_result VARCHAR(50) NOT NULL DEFAULT '',
    last_run_message TEXT NOT NULL DEFAULT '',
    next_run_at TIMESTAMPTZ,
    run_count INT NOT NULL DEFAULT 0,
    max_retries INT NOT NULL DEFAULT 0,
    retry_count INT NOT NULL DEFAULT 0,
    timeout_seconds INT NOT NULL DEFAULT 300,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_scheduled_tasks_enabled ON scheduled_tasks(enabled);