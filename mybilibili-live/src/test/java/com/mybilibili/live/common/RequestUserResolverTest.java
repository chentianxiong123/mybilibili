package com.mybilibili.live.common;

import com.mybilibili.common.exception.BusinessException;
import jakarta.servlet.http.HttpServletRequest;
import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

class RequestUserResolverTest {

    private final RequestUserResolver resolver = new RequestUserResolver();

    @Test
    void requireUserResolvesGatewayUserHeaders() {
        HttpServletRequest request = mock(HttpServletRequest.class);
        when(request.getHeader("X-User-Id")).thenReturn("12");
        when(request.getHeader("X-Username")).thenReturn("alice");

        AuthUser user = resolver.requireUser(request);

        assertEquals(12L, user.id());
        assertEquals("alice", user.username());
    }

    @Test
    void requireAdminResolvesAdminHeaderAndDefaultsName() {
        HttpServletRequest request = mock(HttpServletRequest.class);
        when(request.getHeader("X-Admin-Id")).thenReturn("7");

        AuthUser admin = resolver.requireAdmin(request);

        assertEquals(7L, admin.id());
        assertEquals("管理员7", admin.username());
    }

    @Test
    void requireUserRejectsMissingHeader() {
        HttpServletRequest request = mock(HttpServletRequest.class);

        BusinessException ex = assertThrows(BusinessException.class, () -> resolver.requireUser(request));

        assertEquals(401, ex.getCode());
        assertEquals("未登录", ex.getMessage());
    }

    @Test
    void requireAdminRejectsInvalidHeader() {
        HttpServletRequest request = mock(HttpServletRequest.class);
        when(request.getHeader("X-Admin-Id")).thenReturn("-1");

        BusinessException ex = assertThrows(BusinessException.class, () -> resolver.requireAdmin(request));

        assertEquals(401, ex.getCode());
        assertEquals("登录态无效", ex.getMessage());
    }
}
