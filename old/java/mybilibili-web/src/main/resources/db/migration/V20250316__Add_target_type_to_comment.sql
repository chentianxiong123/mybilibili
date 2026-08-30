-- 评论系统复用改造：添加 target_type 和 target_id 字段

-- 1. 添加新字段
ALTER TABLE comments
ADD COLUMN target_type VARCHAR(20) NULL COMMENT '评论目标类型：VIDEO-视频/DYNAMIC-动态',
ADD COLUMN target_id INT NULL COMMENT '评论目标ID（根据target_type区分是manuscript_id还是dynamic_id）';

-- 2. 迁移现有数据：将现有的 manuscript_id 数据标记为 VIDEO 类型
UPDATE comments
SET target_type = 'VIDEO',
    target_id = manuscript_id
WHERE target_type IS NULL;

-- 3. 设置默认值和约束（可选，根据实际需求决定）
-- ALTER TABLE comments
-- MODIFY COLUMN target_type VARCHAR(20) NOT NULL DEFAULT 'VIDEO',
-- MODIFY COLUMN target_id INT NOT NULL;

-- 4. 创建索引优化查询性能
CREATE INDEX idx_comments_target ON comments(target_type, target_id);
CREATE INDEX idx_comments_target_created ON comments(target_type, target_id, created_at);
