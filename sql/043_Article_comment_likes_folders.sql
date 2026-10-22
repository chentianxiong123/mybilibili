-- 专栏评论点赞
CREATE TABLE IF NOT EXISTS article_comment_likes (
    comment_id BIGINT NOT NULL REFERENCES article_comments(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (comment_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_article_comment_likes_user ON article_comment_likes(user_id);

-- 专栏收藏夹归类（与视频收藏夹共用 favorite_folders）
CREATE TABLE IF NOT EXISTS favorite_folder_articles (
    id BIGSERIAL PRIMARY KEY,
    folder_id BIGINT NOT NULL REFERENCES favorite_folders(id) ON DELETE CASCADE,
    article_id BIGINT NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(folder_id, article_id)
);
CREATE INDEX IF NOT EXISTS idx_fav_folder_articles_folder ON favorite_folder_articles(folder_id);
CREATE INDEX IF NOT EXISTS idx_fav_folder_articles_article ON favorite_folder_articles(article_id);
