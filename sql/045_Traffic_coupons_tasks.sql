-- 创作者流量与激励后端：流量券（曝光加权）+ 投稿/活跃任务。
-- 无真实钱款：券是平台内部曝光资源，redeem 后给稿件加权，推荐流据此提升排序概率。
-- 收益(income_records)仍为真实 0，将来接充电打赏再挂钩。

-- 1. 稿件表补曝光权重：redeem 流量券累加，推荐 SQL 据此加权
ALTER TABLE manuscripts ADD COLUMN IF NOT EXISTS exposure_weight BIGINT NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS idx_manuscripts_exposure_weight ON manuscripts(exposure_weight);

-- 2. 流量券发放记录（平台发给创作者，一条=一张券）
CREATE TABLE IF NOT EXISTS traffic_coupons (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,   -- 券归属创作者
    amount INT NOT NULL DEFAULT 0,                                    -- 券面值（曝光点数）
    reason VARCHAR(200) NOT NULL DEFAULT '',                          -- 发放原因（任务奖励/运营/测试）
    task_id BIGINT NOT NULL DEFAULT 0,                                -- 来源任务（0=非任务发放）
    status INT NOT NULL DEFAULT 0,                                    -- 0 未使用 / 1 已使用 / -1 过期作废
    granted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ                                           -- 有效期，NULL=永久
);
CREATE INDEX IF NOT EXISTS idx_traffic_coupons_user ON traffic_coupons(user_id, status);

-- 3. 流量券使用记录：把券的曝光点数投到一个稿件上
CREATE TABLE IF NOT EXISTS traffic_coupon_uses (
    id BIGSERIAL PRIMARY KEY,
    coupon_id BIGINT NOT NULL REFERENCES traffic_coupons(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    manuscript_id BIGINT NOT NULL REFERENCES manuscripts(id) ON DELETE CASCADE,
    amount INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_coupon_uses_manuscript ON traffic_coupon_uses(manuscript_id);
CREATE INDEX IF NOT EXISTS idx_coupon_uses_user ON traffic_coupon_uses(user_id);

-- 4. 任务定义（平台配置）
CREATE TABLE IF NOT EXISTS creator_tasks (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,                -- 任务标识，如 first_upload / weekly_active
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    reward_coupon INT NOT NULL DEFAULT 0,            -- 完成后奖励的曝光点数
    target INT NOT NULL DEFAULT 0,                   -- 达标数值
    period VARCHAR(20) NOT NULL DEFAULT 'once',      -- once / daily / weekly / total
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 5. 任务进度（每用户每任务一条）
CREATE TABLE IF NOT EXISTS creator_task_progress (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    task_id BIGINT NOT NULL REFERENCES creator_tasks(id) ON DELETE CASCADE,
    progress INT NOT NULL DEFAULT 0,
    done BOOLEAN NOT NULL DEFAULT false,
    reward_claimed BOOLEAN NOT NULL DEFAULT false,
    period_start TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, task_id)
);
CREATE INDEX IF NOT EXISTS idx_task_progress_user ON creator_task_progress(user_id);