package com.mybilibili.common.utils;

import io.jsonwebtoken.Claims;
import io.jsonwebtoken.Jwts;
import io.jsonwebtoken.SignatureAlgorithm;
import java.util.Date;

public class JwtUtils {
    // 密钥
    private static final String SECRET_KEY = "mybilibili_secret_key_2026";
    // 过期时间（24小时）
    private static final long EXPIRATION_TIME = 24 * 60 * 60 * 1000;

    // 生成JWT令牌
    public static String generateToken(Integer userId, String username) {
        Date now = new Date();
        Date expiration = new Date(now.getTime() + EXPIRATION_TIME);

        return Jwts.builder()
                .setSubject(userId.toString())
                .claim("username", username)
                .setIssuedAt(now)
                .setExpiration(expiration)
                .signWith(SignatureAlgorithm.HS256, SECRET_KEY)
                .compact();
    }

    // 解析JWT令牌
    public static Claims parseToken(String token) {
        return Jwts.parser()
                .setSigningKey(SECRET_KEY)
                .parseClaimsJws(token)
                .getBody();
    }

    // 从令牌中获取用户ID
    public static Integer getUserIdFromToken(String token) {
        if (token == null || token.trim().isEmpty()) {
            return null;
        }
        // 去除Bearer前缀
        if (token.startsWith("Bearer ")) {
            token = token.substring(7);
        }
        Claims claims = parseToken(token);
        return Integer.parseInt(claims.getSubject());
    }

    // 从令牌中获取用户名
    public static String getUsernameFromToken(String token) {
        Claims claims = parseToken(token);
        return claims.get("username", String.class);
    }

    // 验证令牌是否过期
    public static boolean isTokenExpired(String token) {
        Claims claims = parseToken(token);
        return claims.getExpiration().before(new Date());
    }
    
    // 从请求中获取用户ID
    public static Integer getUserIdFromRequest(javax.servlet.http.HttpServletRequest request) {
        String xUserId = request.getHeader("X-User-Id");
        if (xUserId != null && !xUserId.isEmpty()) {
            return Integer.parseInt(xUserId);
        }
        String token = request.getHeader("Authorization");
        return getUserIdFromToken(token);
    }
}