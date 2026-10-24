-- 专栏审核：对齐视频稿件流程
-- status: 0=草稿 1=已发布/待审 2=审核拒绝 -1=已删除
-- review_status: 0=待审核 1=已通过 2=已拒绝
-- 公开可见条件：status=1 AND review_status=1 AND visibility=1

ALTER TABLE articles
    ADD COLUMN IF NOT EXISTS review_status INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS review_reason VARCHAR(500) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS review_time TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS reviewer_id BIGINT;

-- 存量已发布专栏视为已通过审核，避免迁移后集体从列表消失
UPDATE articles SET review_status = 1 WHERE status = 1 AND review_status = 0;

CREATE INDEX IF NOT EXISTS idx_articles_review ON articles(review_status, status, publish_time DESC);
