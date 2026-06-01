package com.mybilibili.common.config;

import com.mybilibili.common.utils.JwtUtils;
import jakarta.servlet.FilterChain;
import org.junit.jupiter.api.Test;
import org.springframework.mock.web.MockHttpServletRequest;
import org.springframework.mock.web.MockHttpServletResponse;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNull;
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
        assertEquals("未授权，请登录", response.getContentAsString());
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
        verify(chain).doFilter(request, response);
    }

    @Test
    void protectedPathWithInvalidTokenReturnsUnauthorized() throws Exception {
        MockHttpServletRequest request = request("/api/meeting/my-rooms", "Bearer invalid-token");
        MockHttpServletResponse response = new MockHttpServletResponse();
        FilterChain chain = mock(FilterChain.class);

        filter.doFilterInternal(request, response, chain);

        assertEquals(401, response.getStatus());
        assertEquals("token无效", response.getContentAsString());
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
