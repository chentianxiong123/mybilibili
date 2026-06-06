package com.mybilibili.gateway.filter;

import org.junit.jupiter.api.Test;
import org.springframework.http.HttpMethod;

import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

class GatewayRequestPolicyTest {

    private final GatewayRequestPolicy policy = new GatewayRequestPolicy();

    @Test
    void classifiesPublicAndProtectedPaths() {
        assertTrue(policy.isPublicPath("/actuator/health"));
        assertTrue(policy.isPublicPath("/actuator/health/readiness"));
        assertTrue(policy.isPublicPath("/api/user/login"));
        assertTrue(policy.isPublicPath("/api/live/room/list"));
        assertTrue(policy.isPublicPath("/api/live/room/100"));
        assertTrue(policy.isPublicPath("/api/search/hot"));

        assertFalse(policy.isPublicPath("/api/meeting/my-rooms"));
        assertFalse(policy.isPublicPath("/api/live/linkmic/apply/1"));
        assertFalse(policy.isPublicPath("/api/admin/live/rooms"));
        assertFalse(policy.isPublicPath("/api/search/admin/index/status"));
        assertFalse(policy.isPublicPath("/api/admin/operation-tasks/list"));
        assertFalse(policy.isPublicPath("/api/admin/audit-logs/list"));
        assertFalse(policy.isPublicPath("/api/creator/stats/overview"));
        assertFalse(policy.isPublicPath("/actuator/prometheus"));
        assertFalse(policy.isPublicPath(HttpMethod.POST, "/api/category"));
        assertFalse(policy.isPublicPath(HttpMethod.POST, "/api/banner-images/home"));
    }

    @Test
    void classifiesAdminAndSuperAdminPaths() {
        assertTrue(policy.isAdminPath("/api/admin/live/rooms"));
        assertTrue(policy.isAdminPath("/api/user/admin/list"));
        assertTrue(policy.isAdminPath("/api/video/admin/list"));
        assertTrue(policy.isAdminPath("/api/video/process/admin/stream"));
        assertTrue(policy.isAdminPath("/api/manuscript/admin/pending"));
        assertTrue(policy.isAdminPath("/api/comment/admin/list"));
        assertTrue(policy.isAdminPath("/api/message/admin/system/broadcast"));
        assertTrue(policy.isAdminPath("/api/search/admin/index/status"));
        assertTrue(policy.isAdminPath("/api/admin/operation-tasks/list"));
        assertTrue(policy.isAdminPath("/api/admin/audit-logs/list"));
        assertTrue(policy.isAdminPath("/api/statistics/overview"));
        assertTrue(policy.isAdminPath(HttpMethod.POST, "/api/category"));
        assertTrue(policy.isAdminPath(HttpMethod.POST, "/api/banner-images/home"));
        assertTrue(policy.isSuperAdminPath("/api/admin/register"));
        assertTrue(policy.isSuperAdminPath("/api/admin/roles"));

        assertFalse(policy.isAdminPath("/api/creator/stats/overview"));
        assertFalse(policy.isAdminPath("/api/meeting/my-rooms"));
        assertFalse(policy.isSuperAdminPath("/api/admin/live/rooms"));
    }

    @Test
    void mapsAdminPathsToPermissionCodes() {
        assertTrue("user:manage".equals(policy.requiredPermission(HttpMethod.GET, "/api/user/admin/list")));
        assertTrue("video:manage".equals(policy.requiredPermission(HttpMethod.GET, "/api/video/admin/list")));
        assertTrue("video:manage".equals(policy.requiredPermission(HttpMethod.GET, "/api/video/process/admin/stream")));
        assertTrue("review:manage".equals(policy.requiredPermission(HttpMethod.GET, "/api/manuscript/admin/pending")));
        assertTrue("comment:manage".equals(policy.requiredPermission(HttpMethod.GET, "/api/comment/admin/list")));
        assertTrue("search:manage".equals(policy.requiredPermission(HttpMethod.GET, "/api/search/admin/index/status")));
        assertTrue("operation:manage".equals(policy.requiredPermission(HttpMethod.GET, "/api/admin/operation-tasks/list")));
        assertTrue("audit:manage".equals(policy.requiredPermission(HttpMethod.GET, "/api/admin/audit-logs/list")));
        assertTrue("category:manage".equals(policy.requiredPermission(HttpMethod.POST, "/api/category")));
        assertTrue("banner:manage".equals(policy.requiredPermission(HttpMethod.POST, "/api/banner-images/home")));
    }
}
