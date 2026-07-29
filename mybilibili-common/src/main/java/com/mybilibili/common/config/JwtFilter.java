package com.mybilibili.common.config;

import com.mybilibili.common.security.JwtPrincipal;
import com.mybilibili.common.utils.JwtUtils;
import io.jsonwebtoken.Claims;
import org.springframework.stereotype.Component;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.authority.SimpleGrantedAuthority;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.web.filter.OncePerRequestFilter;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletRequestWrapper;
import jakarta.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.Collection;
import java.util.Collections;
import java.util.Enumeration;
import java.util.List;

@Component
public class JwtFilter extends OncePerRequestFilter {

    private final JwtRequestPolicy requestPolicy = new JwtRequestPolicy();

    @Override
    protected void doFilterInternal(HttpServletRequest request, HttpServletResponse response, FilterChain filterChain) throws ServletException, IOException {
        String path = request.getRequestURI();
        if (requestPolicy.isPublicPath(request.getMethod(), path)) {
            filterChain.doFilter(request, response);
            return;
        }

        String authorization = request.getHeader("Authorization");
        if (authorization == null || !authorization.startsWith("Bearer ")) {
            writeUnauthorized(response, "未授权，请登录");
            return;
        }

        String token = authorization.substring(7);
        UsernamePasswordAuthenticationToken authentication;
        Integer authenticatedUserId;
        String authenticatedUsername;
        String authenticatedRole;
        String tokenType;
        try {
            if (JwtUtils.isTokenExpired(token)) {
                writeUnauthorized(response, "token已过期");
                return;
            }

            Claims claims = JwtUtils.parseToken(token);
            authenticatedUserId = Integer.parseInt(claims.getSubject());
            authenticatedUsername = claims.get("username", String.class);
            authenticatedRole = claims.get("role", String.class);
            tokenType = claims.get("type", String.class);

            request.setAttribute("userId", authenticatedUserId);
            if ("admin".equals(tokenType)) {
                request.setAttribute("adminId", authenticatedUserId);
            }

            JwtPrincipal principal = new JwtPrincipal(authenticatedUserId, authenticatedUsername, authenticatedRole, tokenType);
            authentication = new UsernamePasswordAuthenticationToken(
                    principal,
                    null,
                    buildAuthorities(authenticatedRole, tokenType, claims.get("permissions"))
            );
        } catch (Exception e) {
            writeUnauthorized(response, "token无效");
            return;
        }

        SecurityContextHolder.getContext().setAuthentication(authentication);
        try {
            filterChain.doFilter(new AuthenticatedHeaderRequest(
                    request,
                    authenticatedUserId,
                    authenticatedUsername,
                    authenticatedRole,
                    tokenType
            ), response);
        } finally {
            SecurityContextHolder.clearContext();
        }
    }

    private List<SimpleGrantedAuthority> buildAuthorities(String role, String tokenType, Object permissionsClaim) {
        List<SimpleGrantedAuthority> authorities = new ArrayList<>();
        authorities.add(new SimpleGrantedAuthority("ROLE_USER"));
        if ("admin".equals(tokenType)) {
            authorities.add(new SimpleGrantedAuthority("ROLE_ADMIN"));
            if ("超级管理员".equals(role)) {
                authorities.add(new SimpleGrantedAuthority("ROLE_SUPER_ADMIN"));
            }
        }
        if (role != null && !role.isBlank()) {
            authorities.add(new SimpleGrantedAuthority("ROLE_NAME_" + role));
        }
        for (String permission : extractPermissions(permissionsClaim)) {
            authorities.add(new SimpleGrantedAuthority("PERMISSION_" + permission));
        }
        return authorities;
    }

    private List<String> extractPermissions(Object permissionsClaim) {
        if (!(permissionsClaim instanceof Collection<?> permissions)) {
            return List.of();
        }
        return permissions.stream()
                .filter(String.class::isInstance)
                .map(String.class::cast)
                .filter(permission -> !permission.isBlank())
                .distinct()
                .toList();
    }

    private void writeUnauthorized(HttpServletResponse response, String message) throws IOException {
        response.setCharacterEncoding(StandardCharsets.UTF_8.name());
        response.setContentType("application/json;charset=UTF-8");
        response.setStatus(HttpServletResponse.SC_UNAUTHORIZED);
        response.getWriter().write("{\"code\":401,\"message\":\"" + message + "\",\"data\":null}");
    }

    private static class AuthenticatedHeaderRequest extends HttpServletRequestWrapper {
        private final Integer userId;
        private final String username;
        private final String role;
        private final String tokenType;

        AuthenticatedHeaderRequest(HttpServletRequest request, Integer userId, String username, String role, String tokenType) {
            super(request);
            this.userId = userId;
            this.username = username;
            this.role = role;
            this.tokenType = tokenType;
        }

        @Override
        public String getHeader(String name) {
            String securityHeader = securityHeader(name);
            if (securityHeader != null) {
                return securityHeader;
            }
            return super.getHeader(name);
        }

        @Override
        public Enumeration<String> getHeaders(String name) {
            String securityHeader = securityHeader(name);
            if (securityHeader != null) {
                return Collections.enumeration(List.of(securityHeader));
            }
            return super.getHeaders(name);
        }

        private String securityHeader(String name) {
            if ("X-User-Id".equalsIgnoreCase(name)) {
                return String.valueOf(userId);
            }
            if ("X-Username".equalsIgnoreCase(name)) {
                return username == null ? "" : username;
            }
            if ("X-User-Role".equalsIgnoreCase(name)) {
                return role == null ? "USER" : role;
            }
            if ("X-Admin-Id".equalsIgnoreCase(name) && "admin".equals(tokenType)) {
                return String.valueOf(userId);
            }
            return null;
        }
    }
}
