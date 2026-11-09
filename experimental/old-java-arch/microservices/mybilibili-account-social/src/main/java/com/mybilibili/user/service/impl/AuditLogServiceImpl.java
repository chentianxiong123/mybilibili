package com.mybilibili.user.service.impl;

import com.mybilibili.common.exception.BusinessException;
import com.mybilibili.user.entity.AuditLog;
import com.mybilibili.user.mapper.AuditLogMapper;
import com.mybilibili.user.service.AuditLogService;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.List;

@Service
public class AuditLogServiceImpl implements AuditLogService {

    @Autowired
    private AuditLogMapper auditLogMapper;

    @Override
    public void record(AuditLog auditLog) {
        if (auditLog == null) {
            throw new BusinessException("审计日志不能为空");
        }
        if (isBlank(auditLog.getModule())) {
            throw new BusinessException("审计模块不能为空");
        }
        if (isBlank(auditLog.getAction())) {
            throw new BusinessException("审计动作不能为空");
        }
        if (auditLog.getResult() == null) {
            throw new BusinessException("审计结果不能为空");
        }
        if (auditLog.getCreatedAt() == null) {
            auditLog.setCreatedAt(LocalDateTime.now());
        }
        auditLogMapper.insert(auditLog);
    }

    @Override
    public void recordFromRequest(HttpServletRequest request, String module, String action, String targetType,
                                  String targetId, Integer result, String message, String detail) {
        AuditLog auditLog = new AuditLog();
        auditLog.setOperatorId(extractOperatorId(request));
        auditLog.setOperatorName(header(request, "X-Username"));
        auditLog.setOperatorRole(header(request, "X-User-Role"));
        auditLog.setModule(module);
        auditLog.setAction(action);
        auditLog.setTargetType(targetType);
        auditLog.setTargetId(targetId);
        auditLog.setRequestMethod(request == null ? null : request.getMethod());
        auditLog.setRequestUri(request == null ? null : request.getRequestURI());
        auditLog.setClientIp(getClientIp(request));
        auditLog.setUserAgent(header(request, "User-Agent"));
        auditLog.setResult(result);
        auditLog.setMessage(limit(message, 255));
        auditLog.setDetail(detail);
        record(auditLog);
    }

    @Override
    public List<AuditLog> getLogsByCondition(String operatorKeyword, String module, String action, Integer result,
                                             String targetKeyword, String startTime, String endTime,
                                             Integer page, Integer size) {
        int normalizedPage = normalizePage(page);
        int normalizedSize = normalizeSize(size);
        int offset = (normalizedPage - 1) * normalizedSize;
        return auditLogMapper.selectByCondition(operatorKeyword, module, action, result, targetKeyword,
                startTime, endTime, offset, normalizedSize);
    }

    @Override
    public Integer countByCondition(String operatorKeyword, String module, String action, Integer result,
                                    String targetKeyword, String startTime, String endTime) {
        return auditLogMapper.countByCondition(operatorKeyword, module, action, result, targetKeyword,
                startTime, endTime);
    }

    @Override
    public AuditLog getById(Long id) {
        if (id == null) {
            throw new BusinessException("审计日志ID不能为空");
        }
        return auditLogMapper.selectById(id);
    }

    private Integer extractOperatorId(HttpServletRequest request) {
        String adminId = header(request, "X-Admin-Id");
        if (!isBlank(adminId)) {
            return parseInteger(adminId, "X-Admin-Id");
        }
        String userId = header(request, "X-User-Id");
        if (!isBlank(userId)) {
            return parseInteger(userId, "X-User-Id");
        }
        return null;
    }

    private Integer parseInteger(String value, String fieldName) {
        try {
            return Integer.parseInt(value);
        } catch (NumberFormatException e) {
            throw new BusinessException(fieldName + "不是合法数字");
        }
    }

    private String header(HttpServletRequest request, String name) {
        if (request == null) {
            return null;
        }
        return request.getHeader(name);
    }

    private String getClientIp(HttpServletRequest request) {
        if (request == null) {
            return null;
        }
        String forwarded = request.getHeader("X-Forwarded-For");
        if (!isBlank(forwarded)) {
            return forwarded.split(",")[0].trim();
        }
        String realIp = request.getHeader("X-Real-IP");
        if (!isBlank(realIp)) {
            return realIp.trim();
        }
        return request.getRemoteAddr();
    }

    private int normalizePage(Integer page) {
        return page == null || page < 1 ? 1 : page;
    }

    private int normalizeSize(Integer size) {
        if (size == null || size < 1) {
            return 10;
        }
        return Math.min(size, 100);
    }

    private String limit(String value, int maxLength) {
        if (value == null || value.length() <= maxLength) {
            return value;
        }
        return value.substring(0, maxLength);
    }

    private boolean isBlank(String value) {
        return value == null || value.isBlank();
    }
}
