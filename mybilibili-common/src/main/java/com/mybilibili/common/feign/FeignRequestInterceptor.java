package com.mybilibili.common.feign;

import feign.RequestInterceptor;
import feign.RequestTemplate;
import org.springframework.stereotype.Component;
import org.springframework.web.context.request.RequestContextHolder;
import org.springframework.web.context.request.ServletRequestAttributes;

import javax.servlet.http.HttpServletRequest;

@Component
public class FeignRequestInterceptor implements RequestInterceptor {

    @Override
    public void apply(RequestTemplate template) {
        ServletRequestAttributes attributes = (ServletRequestAttributes) RequestContextHolder.getRequestAttributes();
        if (attributes != null) {
            HttpServletRequest request = attributes.getRequest();
            String token = request.getHeader("Authorization");
            if (token != null && !token.isEmpty()) {
                template.header("Authorization", token);
                String userId = request.getHeader("X-User-Id");
                if (userId != null && !userId.isEmpty()) {
                    template.header("X-User-Id", userId);
                }
            }
        }
    }
}