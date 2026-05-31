package com.mybilibili.common.config;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

class JwtRequestPolicyTest {

    private final JwtRequestPolicy policy = new JwtRequestPolicy();

    @Test
    void allowsConfiguredPublicRoutesAndAssets() {
        assertTrue(policy.isPublicPath("/api/user/login"));
        assertTrue(policy.isPublicPath("/api/admin/login"));
        assertTrue(policy.isPublicPath("/api/user/42/following"));
        assertTrue(policy.isPublicPath("/api/manuscript/100"));
        assertTrue(policy.isPublicPath("/api/search/hot"));
        assertTrue(policy.isPublicPath("/api/live/room/list"));
        assertTrue(policy.isPublicPath("/api/live/room/100"));
        assertTrue(policy.isPublicPath("GET", "/api/category"));
        assertTrue(policy.isPublicPath("GET", "/api/banner-images/home"));
    }

    @Test
    void protectsAuthenticatedRoutes() {
        assertFalse(policy.isPublicPath("/api/user/following"));
        assertFalse(policy.isPublicPath("/api/user/add-experience"));
        assertFalse(policy.isPublicPath("/api/admin/register"));
        assertFalse(policy.isPublicPath("/api/admin/operation-tasks/list"));
        assertFalse(policy.isPublicPath("/api/admin/audit-logs/list"));
        assertFalse(policy.isPublicPath("/api/meeting/my-rooms"));
        assertFalse(policy.isPublicPath("/api/live/linkmic/apply/1"));
        assertFalse(policy.isPublicPath("/api/manuscript/100/status"));
        assertFalse(policy.isPublicPath("/api/video/admin/list"));
        assertFalse(policy.isPublicPath("/api/search/admin/index/status"));
        assertFalse(policy.isPublicPath("POST", "/api/category"));
        assertFalse(policy.isPublicPath("POST", "/api/banner-images/home"));
        assertFalse(policy.isPublicPath("/uploads/avatar.png"));
        assertFalse(policy.isPublicPath("/api/uploads/avatar.png"));
        assertFalse(policy.isPublicPath("/covers/cover.jpg"));
        assertFalse(policy.isPublicPath("/videos/video.mp4"));
        assertFalse(policy.isPublicPath(null));
        assertFalse(policy.isPublicPath(""));
    }

    @Test
    void classifiesAdminPathsAndPermissions() {
        assertTrue(policy.isAdminPath("GET", "/api/user/admin/list"));
        assertTrue(policy.isAdminPath("GET", "/api/video/admin/list"));
        assertTrue(policy.isAdminPath("GET", "/api/search/admin/index/status"));
        assertTrue(policy.isAdminPath("GET", "/api/admin/operation-tasks/list"));
        assertTrue(policy.isAdminPath("GET", "/api/admin/audit-logs/list"));
        assertTrue(policy.isAdminPath("POST", "/api/category"));
        assertTrue(policy.isAdminPath("POST", "/api/banner-images/home"));
        assertTrue(policy.isSuperAdminPath("POST", "/api/admin/register"));

        assertTrue("user:manage".equals(policy.requiredPermission("GET", "/api/user/admin/list")));
        assertTrue("video:manage".equals(policy.requiredPermission("GET", "/api/video/admin/list")));
        assertTrue("search:manage".equals(policy.requiredPermission("GET", "/api/search/admin/index/status")));
        assertTrue("operation:manage".equals(policy.requiredPermission("GET", "/api/admin/operation-tasks/list")));
        assertTrue("audit:manage".equals(policy.requiredPermission("GET", "/api/admin/audit-logs/list")));
        assertTrue("category:manage".equals(policy.requiredPermission("POST", "/api/category")));
        assertTrue("banner:manage".equals(policy.requiredPermission("POST", "/api/banner-images/home")));
    }
}
