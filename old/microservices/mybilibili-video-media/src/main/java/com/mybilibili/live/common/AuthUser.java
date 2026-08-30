package com.mybilibili.live.common;

public record AuthUser(Long id, String username) {

    public Integer intId() {
        return Math.toIntExact(id);
    }
}
