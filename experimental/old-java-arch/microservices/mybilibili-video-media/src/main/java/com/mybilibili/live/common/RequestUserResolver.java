package com.mybilibili.live.common;

import com.mybilibili.common.exception.BusinessException;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;

@Component
public class RequestUserResolver {

    private static final String USER_ID_HEADER = "X-User-Id";
    private static final String ADMIN_ID_HEADER = "X-Admin-Id";
    private static final String USERNAME_HEADER = "X-Username";

    public AuthUser requireUser(HttpServletRequest request) {
        return resolveUser(request, USER_ID_HEADER, "用户");
    }

    public AuthUser requireAdmin(HttpServletRequest request) {
        return resolveUser(request, ADMIN_ID_HEADER, "管理员");
    }

    private AuthUser resolveUser(HttpServletRequest request, String userIdHeader, String defaultUsernamePrefix) {
        String rawUserId = request.getHeader(userIdHeader);
        if (!StringUtils.hasText(rawUserId)) {
            throw new BusinessException(401, "未登录");
        }

        try {
            Long userId = Long.parseLong(rawUserId);
            if (userId <= 0) {
                throw new NumberFormatException("non-positive user id");
            }
            String username = request.getHeader(USERNAME_HEADER);
            if (!StringUtils.hasText(username)) {
                username = defaultUsernamePrefix + userId;
            }
            return new AuthUser(userId, username);
        } catch (NumberFormatException e) {
            throw new BusinessException(401, "登录态无效");
        }
    }
}
