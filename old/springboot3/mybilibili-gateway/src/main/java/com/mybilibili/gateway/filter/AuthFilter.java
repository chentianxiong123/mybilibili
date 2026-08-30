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
import java.util.Arrays;
import java.util.List;

@Slf4j
@Component
public class AuthFilter implements GlobalFilter, Ordered {

    private static final String SECRET_KEY = "REDACTED_JWT_SECRET";
    private static final SecretKey KEY = Keys.hmacShaKeyFor(SECRET_KEY.getBytes(StandardCharsets.UTF_8));

    private static final List<String> WHITELIST = Arrays.asList(
            "/api/user/login",
            "/api/user/register",
            "/api/user/check",
            "/api/admin/login",
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
            // 直播（无需登录）
            "/api/live/room/list",
            "/api/live/room/",
            "/api/live/room/my",
            // 会议（无需登录）
            "/api/meeting/create",
            "/api/meeting/room/",
            "/api/meeting/join/",
            "/api/meeting/my-rooms",
            "/api/meeting/participants/",
            // 连麦（无需登录）
            "/api/live/linkmic/apply/",
            "/api/live/linkmic/active/",
            "/api/live/linkmic/pending/"
    );

    private static final List<String> SUPER_ADMIN_PATHS = Arrays.asList(
            "/api/admin/roles",
            "/api/admin/permissions",
            "/api/admin/admins"
    );

    private static final List<String> PUBLIC_PATHS = Arrays.asList(
            "/api/user/",
            "/api/video/",
            "/api/manuscript/"
    );

    private static final List<String> AUTH_REQUIRED_PATHS = Arrays.asList(
            "/api/user/following",
            "/api/user/followers",
            "/api/user/update",
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

            // 管理后台路径权限校验
            if (path.startsWith("/api/admin/") || path.startsWith("/api/ai/admin")) {
                String tokenType = claims.get("type", String.class);
                if (!"admin".equals(tokenType)) {
                    return forbidden(exchange.getResponse(), "非管理员无权访问");
                }
                if (SUPER_ADMIN_PATHS.stream().anyMatch(path::startsWith) && !"超级管理员".equals(role)) {
                    return forbidden(exchange.getResponse(), "权限不足，仅超级管理员可操作");
                }
            }

            log.info("AuthFilter - Path: {}, UserId: {}, Username: {}", path, userId, username);

            String authHeader = exchange.getRequest().getHeaders().getFirst(HttpHeaders.AUTHORIZATION);
            ServerHttpRequest mutatedRequest = request.mutate()
                    .header("X-User-Id", String.valueOf(userId))
                    .header("X-Username", username)
                    .header("X-User-Role", role != null ? role : "USER")
                    .header(HttpHeaders.AUTHORIZATION, authHeader)
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
        // 管理端路径不在此放行（需要登录验证）
        if (path.startsWith("/api/admin/live") || path.startsWith("/api/admin/meeting")) {
            return false;
        }
        // 直播相关路径无需登录
        if (path.startsWith("/api/live/")) {
            return true;
        }
        if (AUTH_REQUIRED_PATHS.stream().anyMatch(path::startsWith)) {
            return false;
        }
        if (WHITELIST.stream().anyMatch(path::startsWith)) {
            return true;
        }
        return PUBLIC_PATHS.stream().anyMatch(path::startsWith);
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