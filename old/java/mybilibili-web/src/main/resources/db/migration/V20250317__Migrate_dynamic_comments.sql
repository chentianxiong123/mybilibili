-- 迁移动态评论数据到统一的评论表
-- 注意：此脚本应在 V20250316__Add_target_type_to_comment.sql 执行后运行

-- 1. 将 dynamic_comment 数据迁移到 comments 表（一级评论）
INSERT INTO comments (
    user_id,
    content,
    like_count,
    reply_count,
    created_at,
    updated_at,
    target_type,
    target_id,
    manuscript_id
)
SELECT
    dc.user_id,
    dc.content,
    dc.like_count,
    0, -- reply_count 需要单独计算
    dc.created_at,
    dc.created_at,
    'DYNAMIC',
    dc.dynamic_id,
    NULL -- manuscript_id 为空
FROM dynamic_comment dc
WHERE dc.parent_id IS NULL  -- 只迁移一级评论
AND dc.status = 0;  -- 只迁移正常状态的评论

-- 2. 将 dynamic_comment 的回复数据迁移到 reply 表
-- 注意：需要关联到迁移后的 comment_id
INSERT INTO replies (
    comment_id,
    user_id,
    reply_to_user_id,
    content,
    like_count,
    created_at,
    updated_at
)
SELECT
    c.id AS comment_id,
    dc.user_id,
    dc.reply_user_id,
    dc.content,
    dc.like_count,
    dc.created_at,
    dc.created_at
FROM dynamic_comment dc
JOIN comments c ON c.target_id = dc.dynamic_id
    AND c.target_type = 'DYNAMIC'
    AND c.user_id = dc.parent_id  -- 通过 parent_id 关联到父评论
WHERE dc.parent_id IS NOT NULL  -- 只迁移回复
AND dc.status = 0;

-- 3. 更新动态的评论数统计（可选，用于验证）
-- SELECT dynamic_id, COUNT(*) as comment_count
-- FROM dynamic_comment
-- WHERE parent_id IS NULL AND status = 0
-- GROUP BY dynamic_id;

-- 4. 验证迁移结果
-- SELECT target_type, COUNT(*) as count FROM comments GROUP BY target_type;
