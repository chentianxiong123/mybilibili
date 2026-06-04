package com.mybilibili.gateway.security;

import io.jsonwebtoken.Claims;
import io.jsonwebtoken.Jwts;
import io.jsonwebtoken.security.Keys;
import org.springframework.security.authentication.BadCredentialsException;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.Authentication;
import org.springframework.security.core.authority.SimpleGrantedAuthority;
import org.springframework.security.authentication.ReactiveAuthenticationManager;
import org.springframework.stereotype.Component;
import reactor.core.publisher.Mono;

import javax.crypto.SecretKey;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.Collection;
import java.util.List;

@Component
public class JwtReactiveAuthenticationManager implements ReactiveAuthenticationManager {

    private static final String LOCAL_DEV_SECRET = "mybilibili-local-dev-jwt-secret-0123456789abcdef";
    private static final SecretKey KEY = resolveKey();

    @Override
    public Mono<Authentication> authenticate(Authentication authentication) {
        String token = String.valueOf(authentication.getCredentials());
        try {
            Claims claims = Jwts.parser()
                    .verifyWith(KEY)
                    .build()
                    .parseSignedClaims(token)
                    .getPayload();

            Integer userId = Integer.parseInt(claims.getSubject());
            String username = claims.get("username", String.class);
            String role = claims.get("role", String.class);
            String tokenType = claims.get("type", String.class);
            GatewayUserPrincipal principal = new GatewayUserPrincipal(userId, username, role, tokenType);

            List<SimpleGrantedAuthority> authorities = new ArrayList<>();
            authorities.add(new SimpleGrantedAuthority("ROLE_USER"));
            if (principal.isAdmin()) {
                authorities.add(new SimpleGrantedAuthority("ROLE_ADMIN"));
                if ("超级管理员".equals(role)) {
                    authorities.add(new SimpleGrantedAuthority("ROLE_SUPER_ADMIN"));
                }
            }
            if (role != null && !role.isBlank()) {
                authorities.add(new SimpleGrantedAuthority("ROLE_NAME_" + role));
            }
            for (String permission : extractPermissions(claims.get("permissions"))) {
                authorities.add(new SimpleGrantedAuthority("PERMISSION_" + permission));
            }

            return Mono.just(new UsernamePasswordAuthenticationToken(principal, token, authorities));
        } catch (Exception e) {
            return Mono.error(new BadCredentialsException("Token无效或已过期", e));
        }
    }

    private List<String> extractPermissions(Object permissionsClaim) {
        if (!(permissionsClaim instanceof Collection<?> permissions)) {
            return List.of();
        }
        return permissions.stream()
                .filter(String.class::isInstance)
                .map(String.class::cast)
                .filter(permission -> !permission.isBlank())
                .distinct()
                .toList();
    }

    private static SecretKey resolveKey() {
        String secret = firstNonBlank(System.getProperty("jwt.secret"), System.getenv("JWT_SECRET"), LOCAL_DEV_SECRET);
        if (secret.getBytes(StandardCharsets.UTF_8).length < 32) {
            throw new IllegalStateException("JWT secret must be at least 32 bytes");
        }
        return Keys.hmacShaKeyFor(secret.getBytes(StandardCharsets.UTF_8));
    }

    private static String firstNonBlank(String... values) {
        for (String value : values) {
            if (value != null && !value.isBlank()) {
                return value;
            }
        }
        return LOCAL_DEV_SECRET;
    }

}
