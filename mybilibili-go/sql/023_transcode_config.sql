CREATE TABLE IF NOT EXISTS system_configs (
    config_key   VARCHAR(100) PRIMARY KEY,
    config_value TEXT NOT NULL DEFAULT '',
    updated_at   TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_by   VARCHAR(100) NOT NULL DEFAULT ''
);

INSERT INTO system_configs (config_key, config_value, updated_by)
VALUES ('transcode_encoder', 'auto', 'system')
ON CONFLICT (config_key) DO NOTHING;