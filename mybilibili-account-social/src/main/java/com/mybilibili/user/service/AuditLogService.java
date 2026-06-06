package com.mybilibili.user.service;

import com.mybilibili.user.entity.AuditLog;
import jakarta.servlet.http.HttpServletRequest;

import java.util.List;

public interface AuditLogService {
    void record(AuditLog auditLog);

    void recordFromRequest(HttpServletRequest request, String module, String action, String targetType,
                           String targetId, Integer result, String message, String detail);

    List<AuditLog> getLogsByCondition(String operatorKeyword, String module, String action, Integer result,
                                      String targetKeyword, String startTime, String endTime,
                                      Integer page, Integer size);

    Integer countByCondition(String operatorKeyword, String module, String action, Integer result,
                             String targetKeyword, String startTime, String endTime);

    AuditLog getById(Long id);
}
