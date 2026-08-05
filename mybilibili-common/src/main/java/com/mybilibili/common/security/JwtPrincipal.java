package com.mybilibili.common.security;

public record JwtPrincipal(Integer userId, String username, String role, String tokenType) {
}
