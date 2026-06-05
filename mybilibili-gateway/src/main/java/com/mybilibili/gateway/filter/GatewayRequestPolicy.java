package com.mybilibili.gateway.filter;

import org.springframework.stereotype.Component;
import org.springframework.http.HttpMethod;

import java.util.List;
import java.util.regex.Pattern;

@Component
public final class GatewayRequestPolicy {

    private static final List<String> PUBLIC_GET_PATH_PREFIXES = List.of(
            "/actuator/health",
            "/actuator/info",
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
            "/api/ai/summary/",
            "/api/live/room/list"
    );

    private static final List<String> PUBLIC_EXACT_PATHS = List.of(
            "/api/user/login",
            "/api/user/register",
            "/api/user/check",
            "/api/user/email/code",
            "/api/user/email/verify",
            "/api/user/password/forgot",
            "/api/user/default-avatar",
            "/api/admin/login",
            "/api/live/room/srs/hook",
            "/ws/notification",
            "/ws/danmaku",
            "/api/user/token/refresh"
    );

    private static final List<String> ADMIN_PATH_PREFIXES = List.of(
            "/api/admin/",
            "/api/ai/admin",
            "/api/ai/process/",
            "/api/comment/admin/",
            "/api/manuscript/admin/",
            "/manuscript/admin/",
            "/api/message/admin/",
            "/api/search/admin/",
            "/api/user/admin/",
            "/api/video/admin/",
            "/api/video/process/admin/",
            "/api/statistics/"
    );

    private static final List<String> SUPER_ADMIN_PATH_PREFIXES = List.of(
            "/api/admin/register",
            "/api/admin/roles",
            "/api/admin/permissions",
            "/api/admin/admins"
    );

    private static final List<PermissionRule> PERMISSION_RULES = List.of(
            new PermissionRule("/api/admin/roles", "role:manage"),
            new PermissionRule("/api/admin/permissions", "role:manage"),
            new PermissionRule("/api/admin/admins", "admin:manage"),
            new PermissionRule("/api/admin/register", "admin:manage"),
            new PermissionRule("/api/admin/operation-tasks/", "operation:manage"),
            new PermissionRule("/api/admin/audit-logs/", "audit:manage"),
            new PermissionRule("/api/admin/login-logs/", "security:manage"),
            new PermissionRule("/api/admin/content-review/", "review:manage"),
            new PermissionRule("/api/admin/report/", "review:manage"),
            new PermissionRule("/api/admin/prohibited-words/", "comment:manage"),
            new PermissionRule("/api/admin/security-settings/", "security:manage"),
            new PermissionRule("/api/admin/live/", "live:manage"),
            new PermissionRule("/api/admin/meeting/", "meeting:manage"),
            new PermissionRule("/api/admin/storage/", "storage:manage"),
            new PermissionRule("/api/admin/", "admin:manage"),
            new PermissionRule("/api/ai/admin", "ai:manage"),
            new PermissionRule("/api/ai/process/", "ai:manage"),
            new PermissionRule("/api/comment/admin/", "comment:manage"),
            new PermissionRule("/api/manuscript/admin/", "review:manage"),
            new PermissionRule("/manuscript/admin/", "review:manage"),
            new PermissionRule("/api/message/admin/", "message:manage"),
            new PermissionRule("/api/search/admin/", "search:manage"),
            new PermissionRule("/api/user/admin/", "user:manage"),
            new PermissionRule("/api/video/admin/", "video:manage"),
            new PermissionRule("/api/video/process/admin/", "video:manage"),
            new PermissionRule("/api/statistics/", "statistics:manage")
    );

    private static final List<String> AUTH_REQUIRED_PATH_PREFIXES = List.of(
            "/api/user/following",
            "/api/user/followers",
            "/api/user/update",
            "/api/user/login-logs",
            "/api/user/privacy",
            "/api/user/pinned-video",
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
            "/api/manuscript/me/",
            "/api/creator/stats/",
            "/api/watch-history",
            "/api/message/conversations",
            "/api/message/send",
            "/api/message/unread",
            "/api/message/replies",
            "/api/message/at",
            "/api/message/likes",
            "/api/message/system",
            "/api/message/settings",
            "/api/live/linkmic/",
            "/api/live/room/create",
            "/api/live/room/my",
            "/api/meeting/",
            "/api/profile/"
    );

    private static final Pattern PUBLIC_USER_PROFILE =
            Pattern.compile("^/api/user/\\d+$|^/api/user/\\d+/pinned-video$");
    private static final Pattern PUBLIC_LIVE_READ =
            Pattern.compile("^/api/live/room/\\d+$");

    public boolean isPublicPath(String path) {
        return isPublicPath(HttpMethod.GET, path);
    }

    public boolean isPublicPath(HttpMethod method, String path) {
        if (PUBLIC_EXACT_PATHS.contains(path)) {
            return true;
        }
        if (isAuthRequiredPath(path) || isAdminPath(method, path)) {
            return false;
        }
        if (!HttpMethod.GET.equals(method)) {
            return false;
        }
        if (PUBLIC_USER_PROFILE.matcher(path).matches() || PUBLIC_LIVE_READ.matcher(path).matches()) {
            return true;
        }
        return PUBLIC_GET_PATH_PREFIXES.stream().anyMatch(prefix -> matchesPrefix(path, prefix));
    }

    public boolean isAdminPath(String path) {
        return isAdminPath(HttpMethod.GET, path);
    }

    public boolean isAdminPath(HttpMethod method, String path) {
        return ADMIN_PATH_PREFIXES.stream().anyMatch(prefix -> matchesPrefix(path, prefix))
                || isProtectedCatalogMutation(method, path)
                || isProtectedBannerMutation(method, path);
    }

    public boolean isSuperAdminPath(String path) {
        return isSuperAdminPath(HttpMethod.GET, path);
    }

    public boolean isSuperAdminPath(HttpMethod method, String path) {
        return SUPER_ADMIN_PATH_PREFIXES.stream().anyMatch(prefix -> matchesPrefix(path, prefix));
    }

    public boolean isAuthRequiredPath(String path) {
        return AUTH_REQUIRED_PATH_PREFIXES.stream().anyMatch(prefix -> matchesPrefix(path, prefix));
    }

    public String requiredPermission(HttpMethod method, String path) {
        if (isProtectedCatalogMutation(method, path)) {
            return "category:manage";
        }
        if (isProtectedBannerMutation(method, path)) {
            return "banner:manage";
        }
        return PERMISSION_RULES.stream()
                .filter(rule -> matchesPrefix(path, rule.pathPrefix()))
                .map(PermissionRule::permission)
                .findFirst()
                .orElse(null);
    }

    private boolean isProtectedCatalogMutation(HttpMethod method, String path) {
        return !HttpMethod.GET.equals(method) && matchesPrefix(path, "/api/category");
    }

    private boolean isProtectedBannerMutation(HttpMethod method, String path) {
        return !HttpMethod.GET.equals(method) && matchesPrefix(path, "/api/banner-images/");
    }

    private boolean matchesPrefix(String path, String prefix) {
        if (path == null) {
            return false;
        }
        if (prefix.endsWith("/")) {
            return path.startsWith(prefix);
        }
        return path.equals(prefix) || path.startsWith(prefix + "/");
    }

    private record PermissionRule(String pathPrefix, String permission) {
    }
}
