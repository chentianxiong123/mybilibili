-- 空间访客：记录个人空间被访问（登录用户记 visitor_user_id，未登录记 IP），
-- 同一访客同一天只记一条，供数据中心「空间访客」指标使用。
CREATE TABLE IF NOT EXISTS space_visits (
    id BIGSERIAL PRIMARY KEY,
    host_user_id BIGINT NOT NULL,
    visitor_user_id BIGINT,
    visitor_ip VARCHAR(64) NOT NULL DEFAULT '',
    visit_date DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_space_visits_dedupe
    ON space_visits(host_user_id, visit_date, COALESCE(visitor_user_id, 0), COALESCE(visitor_ip, ''));
CREATE INDEX IF NOT EXISTS idx_space_visits_host ON space_visits(host_user_id, visit_date);
