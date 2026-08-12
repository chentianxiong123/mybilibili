CREATE TABLE IF NOT EXISTS follows (
    id BIGSERIAL PRIMARY KEY,
    follower_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    following_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(follower_id, following_id)
);

CREATE INDEX idx_follows_follower ON follows(follower_id);
CREATE INDEX idx_follows_following ON follows(following_id);

CREATE TABLE IF NOT EXISTS watch_history (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    manuscript_id BIGINT NOT NULL REFERENCES manuscripts(id) ON DELETE CASCADE,
    progress_seconds INT NOT NULL DEFAULT 0,
    watched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, manuscript_id)
);

CREATE INDEX idx_watch_history_user ON watch_history(user_id, watched_at DESC);

CREATE TABLE IF NOT EXISTS favorite_folders (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_favorite_folders_user ON favorite_folders(user_id);

CREATE TABLE IF NOT EXISTS favorite_folder_videos (
    id BIGSERIAL PRIMARY KEY,
    folder_id BIGINT NOT NULL REFERENCES favorite_folders(id) ON DELETE CASCADE,
    manuscript_id BIGINT NOT NULL REFERENCES manuscripts(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(folder_id, manuscript_id)
);

CREATE INDEX idx_fav_folder_videos_folder ON favorite_folder_videos(folder_id);
CREATE INDEX idx_fav_folder_videos_manuscript ON favorite_folder_videos(manuscript_id);