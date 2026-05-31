package com.mybilibili.gateway.security;

import com.mybilibili.gateway.filter.GatewayRequestPolicy;
import org.junit.jupiter.api.Test;
import org.springframework.http.HttpMethod;
import org.springframework.mock.http.server.reactive.MockServerHttpRequest;
import org.springframework.mock.web.server.MockServerWebExchange;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.authority.SimpleGrantedAuthority;
import org.springframework.security.web.server.authorization.AuthorizationContext;
import reactor.core.publisher.Mono;

import java.util.List;

import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

class GatewayAuthorizationManagerTest {

    private final GatewayAuthorizationManager manager = new GatewayAuthorizationManager(new GatewayRequestPolicy());

    @Test
    void allowsPublicPathsWithoutAuthentication() {
        assertTrue(check("/api/search/hot", Mono.empty()));
    }

    @Test
    void deniesProtectedPathsWithoutAuthentication() {
        assertFalse(check("/api/meeting/my-rooms", Mono.empty()));
    }

    @Test
    void deniesUserTokenForAdminPath() {
        var authentication = new UsernamePasswordAuthenticationToken(
                new GatewayUserPrincipal(12, "alice", null, null),
                null,
                List.of(new SimpleGrantedAuthority("ROLE_USER"))
        );

        assertFalse(check("/api/admin/live/rooms", Mono.just(authentication)));
        assertFalse(check("/api/user/admin/list", Mono.just(authentication)));
        assertFalse(check("/api/search/admin/index/status", Mono.just(authentication)));
    }

    @Test
    void allowsAdminAndSuperAdminPathsByAuthority() {
        var admin = new UsernamePasswordAuthenticationToken(
                new GatewayUserPrincipal(7, "boss", "普通管理员", "admin"),
                null,
                List.of(
                        new SimpleGrantedAuthority("ROLE_USER"),
                        new SimpleGrantedAuthority("ROLE_ADMIN"),
                        new SimpleGrantedAuthority("PERMISSION_live:manage")
                )
        );
        var superAdmin = new UsernamePasswordAuthenticationToken(
                new GatewayUserPrincipal(1, "root", "超级管理员", "admin"),
                null,
                List.of(
                        new SimpleGrantedAuthority("ROLE_USER"),
                        new SimpleGrantedAuthority("ROLE_ADMIN"),
                        new SimpleGrantedAuthority("ROLE_SUPER_ADMIN")
                )
        );

        assertTrue(check("/api/admin/live/rooms", Mono.just(admin)));
        assertFalse(check("/api/admin/roles", Mono.just(admin)));
        assertTrue(check("/api/admin/roles", Mono.just(superAdmin)));
    }

    @Test
    void requiresPermissionCodeForAdminPath() {
        var userAdmin = admin("user:manage");
        var videoAdmin = admin("video:manage");

        assertTrue(check("/api/user/admin/list", Mono.just(userAdmin)));
        assertFalse(check("/api/video/admin/list", Mono.just(userAdmin)));
        assertTrue(check("/api/video/admin/list", Mono.just(videoAdmin)));
    }

    @Test
    void protectsCatalogAndBannerMutationsByMethod() {
        assertTrue(check(HttpMethod.GET, "/api/category", Mono.empty()));
        assertFalse(check(HttpMethod.POST, "/api/category", Mono.empty()));
        assertTrue(check(HttpMethod.POST, "/api/category", Mono.just(admin("category:manage"))));

        assertTrue(check(HttpMethod.GET, "/api/banner-images/home", Mono.empty()));
        assertFalse(check(HttpMethod.POST, "/api/banner-images/home", Mono.just(admin("category:manage"))));
        assertTrue(check(HttpMethod.POST, "/api/banner-images/home", Mono.just(admin("banner:manage"))));
    }

    private boolean check(String path, Mono<org.springframework.security.core.Authentication> authentication) {
        return check(HttpMethod.GET, path, authentication);
    }

    private boolean check(HttpMethod method, String path, Mono<org.springframework.security.core.Authentication> authentication) {
        var exchange = MockServerWebExchange.from(MockServerHttpRequest.method(method, path).build());
        return manager.check(authentication, new AuthorizationContext(exchange)).block().isGranted();
    }

    private UsernamePasswordAuthenticationToken admin(String permission) {
        return new UsernamePasswordAuthenticationToken(
                new GatewayUserPrincipal(7, "boss", "普通管理员", "admin"),
                null,
                List.of(
                        new SimpleGrantedAuthority("ROLE_USER"),
                        new SimpleGrantedAuthority("ROLE_ADMIN"),
                        new SimpleGrantedAuthority("PERMISSION_" + permission)
                )
        );
    }
}
