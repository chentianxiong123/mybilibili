CREATE TABLE IF NOT EXISTS recommend_configs (
    id BIGSERIAL PRIMARY KEY,
    config_key VARCHAR(50) UNIQUE NOT NULL,
    config_json TEXT NOT NULL DEFAULT '{}',
    updated_by VARCHAR(100) NOT NULL DEFAULT 'admin',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO recommend_configs (config_key, config_json) VALUES ('default', '{"refresh_interval":300,"for_you_size":20,"related_size":10,"hot_size":10,"personalized":true}')
ON CONFLICT (config_key) DO NOTHING;
