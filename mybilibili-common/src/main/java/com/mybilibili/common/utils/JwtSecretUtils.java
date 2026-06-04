package com.mybilibili.common.utils;

import io.jsonwebtoken.security.Keys;

import javax.crypto.SecretKey;
import java.nio.charset.StandardCharsets;

public final class JwtSecretUtils {

    private static final int MIN_SECRET_BYTES = 32;
    private static final String LOCAL_DEV_SECRET = "mybilibili-local-dev-jwt-secret-0123456789abcdef";

    private JwtSecretUtils() {
    }

    public static SecretKey resolveKey() {
        return Keys.hmacShaKeyFor(resolveSecret().getBytes(StandardCharsets.UTF_8));
    }

    public static String resolveSecret() {
        String secret = firstNonBlank(
                System.getProperty("jwt.secret"),
                System.getenv("JWT_SECRET"),
                LOCAL_DEV_SECRET
        );
        if (secret == null || secret.isBlank()) {
            throw new IllegalStateException("JWT secret is required: set -Djwt.secret or JWT_SECRET");
        }
        if (secret.getBytes(StandardCharsets.UTF_8).length < MIN_SECRET_BYTES) {
            throw new IllegalStateException("JWT secret must be at least 32 bytes");
        }
        return secret;
    }

    private static String firstNonBlank(String... values) {
        for (String value : values) {
            if (value != null && !value.isBlank()) {
                return value;
            }
        }
        return null;
    }
}
