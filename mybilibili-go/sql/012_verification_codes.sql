CREATE TABLE IF NOT EXISTS verification_codes (
    id BIGSERIAL PRIMARY KEY,
    identifier VARCHAR(255) NOT NULL,
    code_type VARCHAR(20) NOT NULL,
    code_value VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_verification_codes_lookup ON verification_codes(identifier, code_type, expires_at DESC);