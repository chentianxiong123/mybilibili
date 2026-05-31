package com.mybilibili.gateway.security;

import io.jsonwebtoken.Jwts;
import io.jsonwebtoken.security.Keys;
import org.junit.jupiter.api.Test;
import org.springframework.security.authentication.BadCredentialsException;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.Authentication;

import javax.crypto.SecretKey;
import java.nio.charset.StandardCharsets;
import java.time.Instant;
import java.util.List;
import java.util.Date;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertInstanceOf;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

class JwtReactiveAuthenticationManagerTest {

    private static final String SECRET_KEY = "REDACTED_JWT_SECRET";
    private static final SecretKey KEY = Keys.hmacShaKeyFor(SECRET_KEY.getBytes(StandardCharsets.UTF_8));

    private final JwtReactiveAuthenticationManager manager = new JwtReactiveAuthenticationManager();

    @Test
    void authenticatesUserToken() {
        Authentication authentication = manager.authenticate(rawAuthentication(userToken(12, "alice"))).block();

        GatewayUserPrincipal principal = assertInstanceOf(GatewayUserPrincipal.class, authentication.getPrincipal());
        assertEquals(12, principal.userId());
        assertEquals("alice", principal.username());
        assertTrue(authentication.getAuthorities().stream().anyMatch(a -> "ROLE_USER".equals(a.getAuthority())));
    }

    @Test
    void authenticatesSuperAdminToken() {
        Authentication authentication = manager.authenticate(rawAuthentication(adminToken(7, "boss", "超级管理员", List.of("role:manage")))).block();

        GatewayUserPrincipal principal = assertInstanceOf(GatewayUserPrincipal.class, authentication.getPrincipal());
        assertEquals(7, principal.userId());
        assertTrue(principal.isAdmin());
        assertTrue(authentication.getAuthorities().stream().anyMatch(a -> "ROLE_ADMIN".equals(a.getAuthority())));
        assertTrue(authentication.getAuthorities().stream().anyMatch(a -> "ROLE_SUPER_ADMIN".equals(a.getAuthority())));
        assertTrue(authentication.getAuthorities().stream().anyMatch(a -> "PERMISSION_role:manage".equals(a.getAuthority())));
    }

    @Test
    void rejectsInvalidToken() {
        assertThrows(BadCredentialsException.class, () -> manager.authenticate(rawAuthentication("bad-token")).block());
    }

    private Authentication rawAuthentication(String token) {
        return new UsernamePasswordAuthenticationToken(token, token);
    }

    private String userToken(Integer userId, String username) {
        return Jwts.builder()
                .subject(String.valueOf(userId))
                .claim("username", username)
                .issuedAt(new Date())
                .expiration(Date.from(Instant.now().plusSeconds(3600)))
                .signWith(KEY)
                .compact();
    }

    private String adminToken(Integer adminId, String username, String role, List<String> permissions) {
        return Jwts.builder()
                .subject(String.valueOf(adminId))
                .claim("username", username)
                .claim("role", role)
                .claim("type", "admin")
                .claim("permissions", permissions)
                .issuedAt(new Date())
                .expiration(Date.from(Instant.now().plusSeconds(3600)))
                .signWith(KEY)
                .compact();
    }
}
