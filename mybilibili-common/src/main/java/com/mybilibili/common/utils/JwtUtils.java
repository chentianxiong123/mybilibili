package com.mybilibili.common.utils;

import io.jsonwebtoken.Claims;
import io.jsonwebtoken.Jwts;

import javax.crypto.SecretKey;
import java.util.Collection;
import java.util.Date;

public class JwtUtils {
    private static final long EXPIRATION_TIME = 2 * 60 * 60 * 1000; // 2 hours
    private static final long REFRESH_EXPIRATION_TIME = 30L * 24 * 60 * 60 * 1000; // 30 days
    private static final SecretKey KEY = JwtSecretUtils.resolveKey();

    public static String generateToken(Integer userId, String username) {
        Date now = new Date();
        Date expiration = new Date(now.getTime() + EXPIRATION_TIME);

        return Jwts.builder()
                .subject(userId.toString())
                .claim("username", username)
                .claim("type", "access")
                .issuedAt(now)
                .expiration(expiration)
                .signWith(KEY)
                .compact();
    }

    public static String generateRefreshToken(Integer userId, String username) {
        Date now = new Date();
        Date expiration = new Date(now.getTime() + REFRESH_EXPIRATION_TIME);

        return Jwts.builder()
                .subject(userId.toString())
                .claim("username", username)
                .claim("type", "refresh")
                .issuedAt(now)
                .expiration(expiration)
                .signWith(KEY)
                .compact();
    }

    public static String generateAdminToken(Integer adminId, String username, String role) {
        return generateAdminToken(adminId, username, role, java.util.List.of());
    }

    public static String generateAdminToken(Integer adminId, String username, String role, Collection<String> permissions) {
        Date now = new Date();
        Date expiration = new Date(now.getTime() + EXPIRATION_TIME);

        return Jwts.builder()
                .subject(adminId.toString())
                .claim("username", username)
                .claim("role", role)
                .claim("type", "admin")
                .claim("permissions", permissions == null ? java.util.List.of() : permissions)
                .issuedAt(now)
                .expiration(expiration)
                .signWith(KEY)
                .compact();
    }

    public static Claims parseToken(String token) {
        return Jwts.parser()
                .verifyWith(KEY)
                .build()
                .parseSignedClaims(token)
                .getPayload();
    }

    public static Integer getUserIdFromToken(String token) {
        if (token == null || token.trim().isEmpty()) {
            return null;
        }
        if (token.startsWith("Bearer ")) {
            token = token.substring(7);
        }
        Claims claims = parseToken(token);
        return Integer.parseInt(claims.getSubject());
    }

    public static String getUsernameFromToken(String token) {
        Claims claims = parseToken(token);
        return claims.get("username", String.class);
    }

    public static boolean isTokenExpired(String token) {
        Claims claims = parseToken(token);
        return claims.getExpiration().before(new Date());
    }

    public static Integer getUserIdFromRequest(jakarta.servlet.http.HttpServletRequest request) {
        String xUserId = request.getHeader("X-User-Id");
        if (xUserId != null && !xUserId.isEmpty()) {
            return Integer.parseInt(xUserId);
        }
        String token = request.getHeader("Authorization");
        return getUserIdFromToken(token);
    }
}
