package com.mybilibili.admin.config;

import com.mybilibili.common.utils.JwtUtils;
import io.jsonwebtoken.Claims;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import javax.servlet.FilterChain;
import javax.servlet.ServletException;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.io.IOException;

@Component
public class AdminJwtFilter extends OncePerRequestFilter {

    @Override
    protected void doFilterInternal(HttpServletRequest request, HttpServletResponse response, FilterChain filterChain) throws ServletException, IOException {
        String path = request.getRequestURI();
        
        // 跳过登录、注册接口、测试接口、Swagger相关接口和静态资源
        if (path.contains("/admin/user/login") || path.contains("/admin/user/register") || path.contains("/test") || path.contains("/swagger") || path.contains("/v3/api-docs") || path.contains("/webjars") || path.contains("/uploads/")) {
            filterChain.doFilter(request, response);
            return;
        }
        
        // 暂时跳过上架下架接口的token校验
        if (path.contains("/manuscript/publish/") || path.contains("/manuscript/unpublish/")) {
            filterChain.doFilter(request, response);
            return;
        }

        String token = request.getHeader("Authorization");
        if (token == null || token.isEmpty()) {
            response.setStatus(HttpServletResponse.SC_UNAUTHORIZED);
            response.setContentType("application/json;charset=UTF-8");
            response.getWriter().write("{\"code\": 401, \"message\": \"未授权\", \"data\": null}");
            return;
        }

        // 去除Bearer前缀
        if (token.startsWith("Bearer ")) {
            token = token.substring(7);
        }

        try {
            Claims claims = JwtUtils.parseToken(token);
            Integer adminId = Integer.parseInt(claims.getSubject());
            if (adminId == null) {
                response.setStatus(HttpServletResponse.SC_UNAUTHORIZED);
                response.getWriter().write("{\"code\": 401, \"message\": \"无效的token\", \"data\": null}");
                return;
            }
            request.setAttribute("adminId", adminId);
            request.setAttribute("username", claims.get("username", String.class));
            
            // 设置Spring Security认证信息
            org.springframework.security.core.Authentication authentication = new org.springframework.security.authentication.UsernamePasswordAuthenticationToken(
                adminId, null, java.util.Collections.emptyList()
            );
            org.springframework.security.core.context.SecurityContextHolder.getContext().setAuthentication(authentication);
            
            filterChain.doFilter(request, response);
        } catch (Exception e) {
            response.setStatus(HttpServletResponse.SC_UNAUTHORIZED);
            response.getWriter().write("{\"code\": 401, \"message\": \"token过期或无效\", \"data\": null}");
        }
    }
}