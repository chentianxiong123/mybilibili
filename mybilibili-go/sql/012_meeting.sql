CREATE TABLE IF NOT EXISTS meeting_room (
    id BIGSERIAL PRIMARY KEY,
    room_name VARCHAR(100) NOT NULL,
    room_code VARCHAR(10) UNIQUE NOT NULL,
    creator_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    creator_name VARCHAR(50) NOT NULL DEFAULT '',
    max_participants INT NOT NULL DEFAULT 5,
    status INT NOT NULL DEFAULT 0,
    start_time TIMESTAMPTZ,
    end_time TIMESTAMPTZ,
    scheduled_start TIMESTAMPTZ,
    scheduled_end TIMESTAMPTZ,
    scheduled_reason VARCHAR(500) NOT NULL DEFAULT '',
    create_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    update_time TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_meeting_room_creator ON meeting_room(creator_id);
CREATE INDEX IF NOT EXISTS idx_meeting_room_code ON meeting_room(room_code);

CREATE TABLE IF NOT EXISTS meeting_participant (
    id BIGSERIAL PRIMARY KEY,
    room_id BIGINT NOT NULL REFERENCES meeting_room(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_name VARCHAR(50) NOT NULL DEFAULT '',
    user_avatar VARCHAR(255) NOT NULL DEFAULT '',
    role INT NOT NULL DEFAULT 0,
    audio_enabled INT NOT NULL DEFAULT 0,
    video_enabled INT NOT NULL DEFAULT 0,
    screen_share_enabled INT NOT NULL DEFAULT 0,
    join_time TIMESTAMPTZ,
    leave_time TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_meeting_participant_room ON meeting_participant(room_id);
CREATE INDEX IF NOT EXISTS idx_meeting_participant_user ON meeting_participant(user_id);