package com.mybilibili.gateway.filter;

import io.jsonwebtoken.Claims;
import io.jsonwebtoken.Jwts;
import io.jsonwebtoken.security.Keys;
import lombok.extern.slf4j.Slf4j;
import org.springframework.cloud.gateway.filter.GatewayFilterChain;
import org.springframework.cloud.gateway.filter.GlobalFilter;
import org.springframework.core.Ordered;
import org.springframework.core.io.buffer.DataBuffer;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.http.server.reactive.ServerHttpRequest;
import org.springframework.http.server.reactive.ServerHttpResponse;
import org.springframework.stereotype.Component;
import org.springframework.web.server.ServerWebExchange;
import reactor.core.publisher.Mono;

import javax.crypto.SecretKey;
import java.nio.charset.StandardCharsets;
import java.util.List;

@Slf4j
@Component
public class AuthFilter implements GlobalFilter, Ordered {

    private static final String SECRET_KEY = "REDACTED_JWT_SECRET";
    private static final SecretKey KEY = Keys.hmacShaKeyFor(SECRET_KEY.getBytes(StandardCharsets.UTF_8));

    private static final List<String> PUBLIC_PATH_PREFIXES = List.of(
            "/api/video/play",
            "/api/video/recommended",
            "/api/video/hot",
            "/api/video/list",
            "/api/video/category",
            "/api/video/manuscript/",
            "/api/video/user/",
            "/api/manuscript/recommended",
            "/api/manuscript/hot",
            "/api/manuscript/list",
            "/api/manuscript/category",
            "/api/search/",
            "/api/search/hot",
            "/api/category",
            "/api/category/",
            "/api/category/list",
            "/api/banner-images/",
            "/api/banner/",
            "/api/recommend/hot",
            "/api/recommend/related/",
            "/api/danmaku/",
            "/api/dynamic/hot",
            "/api/follow/check",
            "/uploads/",
            "/covers/",
            "/videos/",
            "/api/ai/process/",
            "/api/ai/summary/",
            "/api/ai/admin/process/stream",
            "/api/live/",
            "/api/meeting/create",
            "/api/meeting/room/",
            "/api/meeting/join/",
            "/api/meeting/my-rooms",
            "/api/meeting/participants/",
            "/api/live/linkmic/apply/",
            "/api/live/linkmic/active/",
            "/api/live/linkmic/pending/"
    );

    private static final List<String> PUBLIC_EXACT_PATHS = List.of(
            "/api/user/login",
            "/api/user/register",
            "/api/user/check",
            "/api/user/email/code",
            "/api/user/email/verify",
            "/api/user/password/forgot",
            "/api/admin/login"
    );

    private static final List<String> ADMIN_PATH_PREFIXES = List.of(
            "/api/admin/",
            "/api/ai/admin"
    );

    private static final List<String> SUPER_ADMIN_PATH_PREFIXES = List.of(
            "/api/admin/roles",
            "/api/admin/permissions",
            "/api/admin/admins"
    );

    private static final List<String> AUTH_REQUIRED_PATH_PREFIXES = List.of(
            "/api/user/admin/",
            "/api/user/following",
            "/api/user/followers",
            "/api/user/update",
            "/api/user/login-logs",
            "/api/video/favorite",
            "/api/video/like",
            "/api/video/coin",
            "/api/video/collect",
            "/api/video/status",
            "/api/manuscript/favorite",
            "/api/manuscript/like",
            "/api/manuscript/coin",
            "/api/manuscript/collect",
            "/api/manuscript/status",
            "/api/manuscript/share",
            "/api/manuscript/danmaku",
            "/api/manuscript/user/likes",
            "/api/manuscript/user/collections",
            "/api/watch-history",
            "/api/message/conversations",
            "/api/message/send",
            "/api/message/unread",
            "/api/message/replies",
            "/api/message/at",
            "/api/message/likes",
            "/api/message/system",
            "/api/message/settings"
    );

    @Override
    public Mono<Void> filter(ServerWebExchange exchange, GatewayFilterChain chain) {
        ServerHttpRequest request = exchange.getRequest();
        String path = request.getURI().getPath();

        if (isWhitelisted(path)) {
            return chain.filter(exchange);
        }

        String token = getTokenFromRequest(exchange);
        if (token == null) {
            return unauthorized(exchange.getResponse(), "未登录或登录已过期");
        }

        try {
            Claims claims = Jwts.parser()
                    .verifyWith(KEY)
                    .build()
                    .parseSignedClaims(token)
                    .getPayload();

            String userIdStr = claims.getSubject();
            Integer userId = Integer.parseInt(userIdStr);
            String username = claims.get("username", String.class);
            String role = claims.get("role", String.class);

            if (isAdminPath(path)) {
                String tokenType = claims.get("type", String.class);
                if (!"admin".equals(tokenType)) {
                    return forbidden(exchange.getResponse(), "非管理员无权访问");
                }
                if (isSuperAdminPath(path) && !"超级管理员".equals(role)) {
                    return forbidden(exchange.getResponse(), "权限不足，仅超级管理员可操作");
                }
            }

            log.info("AuthFilter - Path: {}, UserId: {}, Username: {}", path, userId, username);

            String authHeader = exchange.getRequest().getHeaders().getFirst(HttpHeaders.AUTHORIZATION);
            ServerHttpRequest mutatedRequest = request.mutate()
                    .headers(headers -> {
                        headers.remove("X-User-Id");
                        headers.remove("X-Admin-Id");
                        headers.remove("X-Username");
                        headers.remove("X-User-Role");
                        headers.set("X-User-Id", String.valueOf(userId));
                        if (isAdminPath(path)) {
                            headers.set("X-Admin-Id", String.valueOf(userId));
                        }
                        headers.set("X-Username", username);
                        headers.set("X-User-Role", role != null ? role : "USER");
                        if (authHeader != null) {
                            headers.set(HttpHeaders.AUTHORIZATION, authHeader);
                        }
                    })
                    .build();

            return chain.filter(exchange.mutate().request(mutatedRequest).build());
        } catch (Exception e) {
            log.error("Token验证失败: {}", e.getMessage());
            return unauthorized(exchange.getResponse(), "Token无效或已过期");
        }
    }

    private String getTokenFromRequest(ServerWebExchange exchange) {
        String authHeader = exchange.getRequest().getHeaders().getFirst(HttpHeaders.AUTHORIZATION);
        if (authHeader != null && authHeader.startsWith("Bearer ")) {
            return authHeader.substring(7);
        }
        return null;
    }

    private boolean isWhitelisted(String path) {
        if (isAuthRequiredPath(path) || isAdminPath(path)) {
            return false;
        }
        if (PUBLIC_EXACT_PATHS.contains(path)) {
            return true;
        }
        return PUBLIC_PATH_PREFIXES.stream().anyMatch(path::startsWith);
    }

    private boolean isAdminPath(String path) {
        return ADMIN_PATH_PREFIXES.stream().anyMatch(path::startsWith);
    }

    private boolean isSuperAdminPath(String path) {
        return SUPER_ADMIN_PATH_PREFIXES.stream().anyMatch(path::startsWith);
    }

    private boolean isAuthRequiredPath(String path) {
        return AUTH_REQUIRED_PATH_PREFIXES.stream().anyMatch(path::startsWith);
    }

    private Mono<Void> unauthorized(ServerHttpResponse response, String message) {
        response.setStatusCode(HttpStatus.UNAUTHORIZED);
        response.getHeaders().setContentType(MediaType.APPLICATION_JSON);
        String body = "{\"code\":401,\"message\":\"" + message + "\",\"data\":null}";
        DataBuffer buffer = response.bufferFactory().wrap(body.getBytes(StandardCharsets.UTF_8));
        return response.writeWith(Mono.just(buffer));
    }

    private Mono<Void> forbidden(ServerHttpResponse response, String message) {
        response.setStatusCode(HttpStatus.FORBIDDEN);
        response.getHeaders().setContentType(MediaType.APPLICATION_JSON);
        String body = "{\"code\":403,\"message\":\"" + message + "\",\"data\":null}";
        DataBuffer buffer = response.bufferFactory().wrap(body.getBytes(StandardCharsets.UTF_8));
        return response.writeWith(Mono.just(buffer));
    }

    @Override
    public int getOrder() {
        return -100;
    }
}