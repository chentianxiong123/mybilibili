-- 数据库重构：迁移交互数据到 user_interactions 表

-- 1. 迁移视频点赞数据（likes 表）
INSERT INTO user_interactions (user_id, target_type, target_id, interaction_type, created_at)
SELECT user_id, 'VIDEO', manuscript_id, 'LIKE', created_at 
FROM likes
WHERE NOT EXISTS (
    SELECT 1 FROM user_interactions ui 
    WHERE ui.user_id = likes.user_id 
    AND ui.target_type = 'VIDEO' 
    AND ui.target_id = likes.manuscript_id 
    AND ui.interaction_type = 'LIKE'
);

-- 2. 迁移动态点赞数据（dynamic_likes 表）
INSERT INTO user_interactions (user_id, target_type, target_id, interaction_type, created_at)
SELECT user_id, 'DYNAMIC', dynamic_id, 'LIKE', created_at 
FROM dynamic_likes
WHERE NOT EXISTS (
    SELECT 1 FROM user_interactions ui 
    WHERE ui.user_id = dynamic_likes.user_id 
    AND ui.target_type = 'DYNAMIC' 
    AND ui.target_id = dynamic_likes.dynamic_id 
    AND ui.interaction_type = 'LIKE'
);

-- 3. 迁移评论点赞数据（dynamic_comment_likes 表）
INSERT INTO user_interactions (user_id, target_type, target_id, interaction_type, created_at)
SELECT user_id, 'COMMENT', comment_id, 'LIKE', created_at 
FROM dynamic_comment_likes
WHERE NOT EXISTS (
    SELECT 1 FROM user_interactions ui 
    WHERE ui.user_id = dynamic_comment_likes.user_id 
    AND ui.target_type = 'COMMENT' 
    AND ui.target_id = dynamic_comment_likes.comment_id 
    AND ui.interaction_type = 'LIKE'
);

-- 4. 迁移收藏数据（collections 表）
INSERT INTO user_interactions (user_id, target_type, target_id, interaction_type, created_at)
SELECT user_id, 'VIDEO', video_id, 'COLLECT', created_at 
FROM collections
WHERE NOT EXISTS (
    SELECT 1 FROM user_interactions ui 
    WHERE ui.user_id = collections.user_id 
    AND ui.target_type = 'VIDEO' 
    AND ui.target_id = collections.video_id 
    AND ui.interaction_type = 'COLLECT'
);

-- 5. 迁移关注数据（follows 表）
INSERT INTO user_interactions (user_id, target_type, target_id, interaction_type, created_at)
SELECT follower_id, 'USER', followed_id, 'FOLLOW', created_at 
FROM follows
WHERE NOT EXISTS (
    SELECT 1 FROM user_interactions ui 
    WHERE ui.user_id = follows.follower_id 
    AND ui.target_type = 'USER' 
    AND ui.target_id = follows.followed_id 
    AND ui.interaction_type = 'FOLLOW'
);

-- 6. 迁移投币数据（coins 表）
INSERT INTO user_interactions (user_id, target_type, target_id, interaction_type, created_at)
SELECT user_id, 'VIDEO', manuscript_id, 'COIN', created_at 
FROM coins
WHERE NOT EXISTS (
    SELECT 1 FROM user_interactions ui 
    WHERE ui.user_id = coins.user_id 
    AND ui.target_type = 'VIDEO' 
    AND ui.target_id = coins.manuscript_id 
    AND ui.interaction_type = 'COIN'
);

-- 7. 初始化交互计数表
-- 点赞计数
INSERT INTO interaction_counts (target_type, target_id, like_count)
SELECT target_type, target_id, COUNT(*) 
FROM user_interactions 
WHERE interaction_type = 'LIKE'
GROUP BY target_type, target_id
ON DUPLICATE KEY UPDATE like_count = VALUES(like_count);

-- 收藏计数
INSERT INTO interaction_counts (target_type, target_id, collect_count)
SELECT target_type, target_id, COUNT(*) 
FROM user_interactions 
WHERE interaction_type = 'COLLECT'
GROUP BY target_type, target_id
ON DUPLICATE KEY UPDATE collect_count = VALUES(collect_count);

-- 投币计数
INSERT INTO interaction_counts (target_type, target_id, coin_count)
SELECT target_type, target_id, COUNT(*) 
FROM user_interactions 
WHERE interaction_type = 'COIN'
GROUP BY target_type, target_id
ON DUPLICATE KEY UPDATE coin_count = VALUES(coin_count);
