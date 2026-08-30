package com.mybilibili.gateway.security;

public record GatewayUserPrincipal(Integer userId, String username, String role, String tokenType) {

    public boolean isAdmin() {
        return "admin".equals(tokenType);
    }
}
