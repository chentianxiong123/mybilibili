package com.mybilibili.common.audit;

import com.mybilibili.common.vo.Result;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.stereotype.Component;

@Component
public class AuditLogRecorder {

    private final JdbcTemplate jdbcTemplate;

    public AuditLogRecorder(JdbcTemplate jdbcTemplate) {
        this.jdbcTemplate = jdbcTemplate;
    }

    public void record(HttpServletRequest request, String module, String action, String targetType,
                       String targetId, Result<?> result) {
        record(request, module, action, targetType, targetId,
                isSuccess(result) ? 1 : 0,
                result == null ? null : result.getMessage(),
                result == null || result.getData() == null ? null : result.getData().toString());
    }

    public void record(HttpServletRequest request, String module, String action, String targetType,
                       String targetId, Integer result, String message, String detail) {
        jdbcTemplate.update("""
                        INSERT INTO audit_logs
                        (operator_id, operator_name, operator_role, module, action, target_type, target_id,
                         request_method, request_uri, client_ip, user_agent, result, message, detail)
                        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                        """,
                operatorId(request),
                header(request, "X-Username"),
                header(request, "X-User-Role"),
                module,
                action,
                targetType,
                targetId,
                request == null ? null : request.getMethod(),
                request == null ? null : request.getRequestURI(),
                clientIp(request),
                header(request, "User-Agent"),
                result,
                limit(message, 255),
                detail);
    }

    private boolean isSuccess(Result<?> result) {
        return result != null && result.getCode() != null && result.getCode() == 200;
    }

    private Integer operatorId(HttpServletRequest request) {
        String adminId = header(request, "X-Admin-Id");
        if (adminId == null || adminId.isBlank()) {
            adminId = header(request, "X-User-Id");
        }
        if (adminId == null || adminId.isBlank()) {
            return null;
        }
        return Integer.parseInt(adminId);
    }

    private String header(HttpServletRequest request, String name) {
        return request == null ? null : request.getHeader(name);
    }

    private String clientIp(HttpServletRequest request) {
        if (request == null) {
            return null;
        }
        String forwarded = request.getHeader("X-Forwarded-For");
        if (forwarded != null && !forwarded.isBlank()) {
            return forwarded.split(",")[0].trim();
        }
        String realIp = request.getHeader("X-Real-IP");
        if (realIp != null && !realIp.isBlank()) {
            return realIp.trim();
        }
        return request.getRemoteAddr();
    }

    private String limit(String value, int maxLength) {
        if (value == null || value.length() <= maxLength) {
            return value;
        }
        return value.substring(0, maxLength);
    }
}
