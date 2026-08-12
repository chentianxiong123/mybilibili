CREATE TABLE live_rooms (
    id BIGSERIAL PRIMARY KEY,
    room_name VARCHAR(100) NOT NULL,
    room_code VARCHAR(10) UNIQUE NOT NULL,
    host_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    stream_key VARCHAR(64) UNIQUE NOT NULL,
    cover VARCHAR(255) NOT NULL DEFAULT '',
    category VARCHAR(50) NOT NULL DEFAULT '',
    max_seats INT NOT NULL DEFAULT 8,
    status INT NOT NULL DEFAULT 0,
    viewer_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_live_rooms_host ON live_rooms(host_id);
CREATE INDEX idx_live_rooms_stream_key ON live_rooms(stream_key);
CREATE INDEX idx_live_rooms_status ON live_rooms(status, viewer_count DESC);

CREATE TABLE live_seats (
    id BIGSERIAL PRIMARY KEY,
    room_id BIGINT NOT NULL REFERENCES live_rooms(id) ON DELETE CASCADE,
    seat_index INT NOT NULL,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    status INT NOT NULL DEFAULT 0,
    muted BOOLEAN NOT NULL DEFAULT FALSE,
    joined_at TIMESTAMPTZ,
    UNIQUE(room_id, seat_index)
);

CREATE INDEX idx_live_seats_user ON live_seats(user_id);

CREATE TABLE live_linkmic (
    id BIGSERIAL PRIMARY KEY,
    room_id BIGINT NOT NULL,
    streamer_id BIGINT NOT NULL,
    viewer_id BIGINT NOT NULL,
    status INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);