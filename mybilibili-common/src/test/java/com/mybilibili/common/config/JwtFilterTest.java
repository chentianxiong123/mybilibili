package com.mybilibili.common.config;

import com.mybilibili.common.utils.JwtUtils;
import jakarta.servlet.FilterChain;
import jakarta.servlet.http.HttpServletRequest;
import org.junit.jupiter.api.Test;
import org.springframework.mock.web.MockHttpServletRequest;
import org.springframework.mock.web.MockHttpServletResponse;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNull;
import org.mockito.ArgumentCaptor;

import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.never;
import static org.mockito.Mockito.verify;

class JwtFilterTest {

    static {
        System.setProperty("jwt.secret", "test-jwt-secret-0123456789abcdef0123456789abcdef");
    }

    private final JwtFilter filter = new JwtFilter();

    @Test
    void publicPathSkipsAuthentication() throws Exception {
        MockHttpServletRequest request = request("/api/user/login", null);
        MockHttpServletResponse response = new MockHttpServletResponse();
        FilterChain chain = mock(FilterChain.class);

        filter.doFilterInternal(request, response, chain);

        assertEquals(200, response.getStatus());
        assertNull(request.getAttribute("userId"));
        verify(chain).doFilter(request, response);
    }

    @Test
    void protectedPathWithoutTokenReturnsUnauthorized() throws Exception {
        MockHttpServletRequest request = request("/api/meeting/my-rooms", null);
        MockHttpServletResponse response = new MockHttpServletResponse();
        FilterChain chain = mock(FilterChain.class);

        filter.doFilterInternal(request, response, chain);

        assertEquals(401, response.getStatus());
        assertEquals("{\"code\":401,\"message\":\"未授权，请登录\",\"data\":null}", response.getContentAsString());
        verify(chain, never()).doFilter(request, response);
    }

    @Test
    void protectedPathWithValidTokenSetsUserId() throws Exception {
        String token = JwtUtils.generateToken(12, "alice");
        MockHttpServletRequest request = request("/api/meeting/my-rooms", "Bearer " + token);
        MockHttpServletResponse response = new MockHttpServletResponse();
        FilterChain chain = mock(FilterChain.class);

        filter.doFilterInternal(request, response, chain);

        assertEquals(200, response.getStatus());
        assertEquals(12, request.getAttribute("userId"));

        ArgumentCaptor<HttpServletRequest> forwardedRequest = ArgumentCaptor.forClass(HttpServletRequest.class);
        verify(chain).doFilter(forwardedRequest.capture(), eq(response));
        assertEquals("12", forwardedRequest.getValue().getHeader("X-User-Id"));
        assertEquals("alice", forwardedRequest.getValue().getHeader("X-Username"));
        assertEquals("USER", forwardedRequest.getValue().getHeader("X-User-Role"));
        assertNull(forwardedRequest.getValue().getHeader("X-Admin-Id"));
    }

    @Test
    void protectedPathWithInvalidTokenReturnsUnauthorized() throws Exception {
        MockHttpServletRequest request = request("/api/meeting/my-rooms", "Bearer invalid-token");
        MockHttpServletResponse response = new MockHttpServletResponse();
        FilterChain chain = mock(FilterChain.class);

        filter.doFilterInternal(request, response, chain);

        assertEquals(401, response.getStatus());
        assertEquals("{\"code\":401,\"message\":\"token无效\",\"data\":null}", response.getContentAsString());
        verify(chain, never()).doFilter(request, response);
    }

    private MockHttpServletRequest request(String uri, String authorization) {
        MockHttpServletRequest request = new MockHttpServletRequest("GET", uri);
        request.setRequestURI(uri);
        if (authorization != null) {
            request.addHeader("Authorization", authorization);
        }
        return request;
    }
}
