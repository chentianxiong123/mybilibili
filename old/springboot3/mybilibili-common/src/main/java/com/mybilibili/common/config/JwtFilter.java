package com.mybilibili.common.config;

import com.mybilibili.common.utils.JwtUtils;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import java.io.IOException;

@Component
public class JwtFilter extends OncePerRequestFilter {

    @Override
    protected void doFilterInternal(HttpServletRequest request, HttpServletResponse response, FilterChain filterChain) throws ServletException, IOException {
        // 获取请求头中的Authorization
        String authorization = request.getHeader("Authorization");
        
        // 不需要认证的路径
        String path = request.getRequestURI();
        // 打印请求路径，用于调试
        System.out.println("Request path: " + path);
        // 静态资源和公开接口不需要认证
        if (path.contains("/user/register") || path.contains("/user/login") ||
            path.contains("/user/add-experience") ||
            path.contains("/admin/register") || path.contains("/admin/login") ||
            path.contains("/user/") && path.matches(".*/user/\\d+(/following|/followers)?$") ||
            path.contains("/swagger") || path.contains("/v3/api-docs") ||
            path.contains("/category") || path.contains("/video/") ||
            path.contains("/follow/") || path.contains("/test/") ||
            path.contains("/covers/") || path.contains("/static/") ||
            path.contains("/images/") || path.contains("/files/") ||
            path.contains("/uploads/") || path.contains("/banner-images/") ||
            path.contains("/recommend/") || path.contains("/search/") ||
            path.contains("/manuscript/recommended") || path.contains("/api/manuscript/recommended") ||
            path.matches(".*/manuscript/\\d+$") ||
            path.endsWith(".jpg") ||
            path.endsWith(".jpeg") || path.endsWith(".png") ||
            path.endsWith(".gif") || path.endsWith(".mp4") ||
            path.endsWith(".avi") || path.endsWith(".mov")) {
            filterChain.doFilter(request, response);
            return;
        }
        
        // 验证token
        if (authorization == null || !authorization.startsWith("Bearer ")) {
            response.setStatus(HttpServletResponse.SC_UNAUTHORIZED);
            response.getWriter().write("未授权，请登录");
            return;
        }
        
        String token = authorization.substring(7);
        try {
            // 验证token是否过期
            if (JwtUtils.isTokenExpired(token)) {
                response.setStatus(HttpServletResponse.SC_UNAUTHORIZED);
                response.getWriter().write("token已过期");
                return;
            }
            
            // 将用户信息存储到请求中
            Integer userId = JwtUtils.getUserIdFromToken(token);
            request.setAttribute("userId", userId);
            
            filterChain.doFilter(request, response);
        } catch (Exception e) {
            response.setStatus(HttpServletResponse.SC_UNAUTHORIZED);
            response.getWriter().write("token无效");
        }
    }
}