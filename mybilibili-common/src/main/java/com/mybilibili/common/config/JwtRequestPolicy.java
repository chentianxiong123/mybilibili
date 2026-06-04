package com.mybilibili.common.config;

import java.util.List;
import java.util.regex.Pattern;

final class JwtRequestPolicy {

    private static final List<String> DISABLED_LOCAL_MEDIA_PREFIXES = List.of(
            "/uploads/",
            "/covers/",
            "/videos/"
    );

    private static final List<String> PUBLIC_EXACT_PATHS = List.of(
            "/user/register",
            "/user/login",
            "/admin/login",
            "/user/email/code",
            "/user/email/verify",
            "/user/password/forgot",
            "/user/default-avatar",
            "/user/token/refresh",
            "/ws/notification",
            "/ws/danmaku"
    );

    private static final List<String> PUBLIC_GET_PREFIXES = List.of(
            "/swagger",
            "/v3/api-docs",
            "/actuator/health",
            "/actuator/info",
            "/category",
            "/video",
            "/follow/check/",
            "/follow/user/",
            "/static/",
            "/images/",
            "/files/",
            "/banner-images/",
            "/live/room/list",
            "/recommend/",
            "/search/",
            "/ai/summary/",
            "/manuscript/recommended",
            "/manuscript/hot",
            "/manuscript/list",
            "/manuscript/category/",
            "/manuscript/user/"
    );

    private static final List<String> ADMIN_PATH_PREFIXES = List.of(
            "/admin/",
            "/ai/admin",
            "/ai/process/",
            "/comment/admin/",
            "/manuscript/admin/",
            "/message/admin/",
            "/search/admin/",
            "/statistics/",
            "/user/admin/",
            "/video/admin/"
    );

    private static final List<String> SUPER_ADMIN_PATH_PREFIXES = List.of(
            "/admin/register",
            "/admin/roles",
            "/admin/permissions",
            "/admin/admins"
    );

    private static final List<PermissionRule> PERMISSION_RULES = List.of(
            new PermissionRule("/admin/roles", "role:manage"),
            new PermissionRule("/admin/permissions", "role:manage"),
            new PermissionRule("/admin/admins", "admin:manage"),
            new PermissionRule("/admin/register", "admin:manage"),
            new PermissionRule("/admin/operation-tasks/", "operation:manage"),
            new PermissionRule("/admin/audit-logs/", "audit:manage"),
            new PermissionRule("/admin/login-logs/", "security:manage"),
            new PermissionRule("/admin/content-review/", "review:manage"),
            new PermissionRule("/admin/report/", "review:manage"),
            new PermissionRule("/admin/prohibited-words/", "comment:manage"),
            new PermissionRule("/admin/security-settings/", "security:manage"),
            new PermissionRule("/admin/live/", "live:manage"),
            new PermissionRule("/admin/meeting/", "meeting:manage"),
            new PermissionRule("/admin/storage/", "storage:manage"),
            new PermissionRule("/admin/", "admin:manage"),
            new PermissionRule("/ai/admin", "ai:manage"),
            new PermissionRule("/ai/process/", "ai:manage"),
            new PermissionRule("/comment/admin/", "comment:manage"),
            new PermissionRule("/manuscript/admin/", "review:manage"),
            new PermissionRule("/message/admin/", "message:manage"),
            new PermissionRule("/search/admin/", "search:manage"),
            new PermissionRule("/statistics/", "statistics:manage"),
            new PermissionRule("/user/admin/", "user:manage"),
            new PermissionRule("/video/admin/", "video:manage")
    );

    private static final List<String> PUBLIC_SUFFIXES = List.of(
            ".jpg",
            ".jpeg",
            ".png",
            ".gif",
            ".mp4",
            ".avi",
            ".mov"
    );

    private static final Pattern PUBLIC_USER_PROFILE =
            Pattern.compile("^/user/\\d+(/following|/followers|/pinned-video)?$");
    private static final Pattern PUBLIC_MANUSCRIPT_DETAIL =
            Pattern.compile("^/manuscript/\\d+$");
    private static final Pattern PUBLIC_LIVE_READ =
            Pattern.compile("^/live/room/\\d+$");

    boolean isPublicPath(String path) {
        return isPublicPath("GET", path);
    }

    boolean isPublicPath(String method, String path) {
        if (path == null || path.isBlank()) {
            return false;
        }
        String normalizedPath = normalize(path);
        if (DISABLED_LOCAL_MEDIA_PREFIXES.stream().anyMatch(path::startsWith)
                || DISABLED_LOCAL_MEDIA_PREFIXES.stream().anyMatch(normalizedPath::startsWith)) {
            return false;
        }
        if (PUBLIC_EXACT_PATHS.contains(normalizedPath)) {
            return true;
        }
        if (isAdminPath(method, normalizedPath)) {
            return false;
        }
        if (!"GET".equalsIgnoreCase(method)) {
            return false;
        }
        return PUBLIC_USER_PROFILE.matcher(normalizedPath).matches()
                || PUBLIC_MANUSCRIPT_DETAIL.matcher(normalizedPath).matches()
                || PUBLIC_LIVE_READ.matcher(normalizedPath).matches()
                || PUBLIC_GET_PREFIXES.stream().anyMatch(prefix -> matchesPrefix(normalizedPath, prefix))
                || hasPublicSuffix(normalizedPath);
    }

    boolean isAdminPath(String method, String path) {
        String normalizedPath = normalize(path);
        return ADMIN_PATH_PREFIXES.stream().anyMatch(prefix -> matchesPrefix(normalizedPath, prefix))
                || isProtectedCatalogMutation(method, normalizedPath)
                || isProtectedBannerMutation(method, normalizedPath);
    }

    boolean isSuperAdminPath(String method, String path) {
        String normalizedPath = normalize(path);
        return SUPER_ADMIN_PATH_PREFIXES.stream().anyMatch(prefix -> matchesPrefix(normalizedPath, prefix));
    }

    String requiredPermission(String method, String path) {
        String normalizedPath = normalize(path);
        if (isProtectedCatalogMutation(method, normalizedPath)) {
            return "category:manage";
        }
        if (isProtectedBannerMutation(method, normalizedPath)) {
            return "banner:manage";
        }
        return PERMISSION_RULES.stream()
                .filter(rule -> matchesPrefix(normalizedPath, rule.pathPrefix()))
                .map(PermissionRule::permission)
                .findFirst()
                .orElse(null);
    }

    private boolean hasPublicSuffix(String path) {
        return PUBLIC_SUFFIXES.stream().anyMatch(path::endsWith);
    }

    private boolean isProtectedCatalogMutation(String method, String path) {
        return !"GET".equalsIgnoreCase(method) && matchesPrefix(path, "/category");
    }

    private boolean isProtectedBannerMutation(String method, String path) {
        return !"GET".equalsIgnoreCase(method) && matchesPrefix(path, "/banner-images/");
    }

    private String normalize(String path) {
        if (path == null) {
            return "";
        }
        if (path.startsWith("/api/")) {
            return path.substring(4);
        }
        return path;
    }

    private boolean matchesPrefix(String path, String prefix) {
        if (prefix.endsWith("/")) {
            return path.startsWith(prefix);
        }
        return path.equals(prefix) || path.startsWith(prefix + "/");
    }

    private record PermissionRule(String pathPrefix, String permission) {
    }
}
