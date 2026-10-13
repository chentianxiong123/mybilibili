-- 稿件申诉：用户对被处理稿件（下架/未通过/锁定）发起申诉，管理员处理后关闭。
-- 前端 /platform/upload-manager/appeal 对齐 B站申诉管理（全部/进行中/已完成 + 申诉卡片）。
CREATE TABLE IF NOT EXISTS appeals (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    manuscript_id BIGINT NOT NULL,
    reason VARCHAR(50) NOT NULL DEFAULT '',
    description VARCHAR(500) NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    admin_remark VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_appeals_user ON appeals(user_id);
CREATE INDEX IF NOT EXISTS idx_appeals_status ON appeals(status);
