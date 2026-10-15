-- 收藏夹可见性：others 打开他人空间只展示公开收藏夹(visible=1)，私密仅本人可见
ALTER TABLE favorite_folders ADD COLUMN IF NOT EXISTS visible SMALLINT NOT NULL DEFAULT 1;
CREATE INDEX IF NOT EXISTS idx_favorite_folders_visible ON favorite_folders(user_id, visible);

-- 存量收藏夹默认视为公开，保证兼容
UPDATE favorite_folders SET visible = 1 WHERE visible = 0;