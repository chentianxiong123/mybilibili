package com.mybilibili.common.operation;

import com.mybilibili.common.vo.Result;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.boot.autoconfigure.condition.ConditionalOnBean;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.stereotype.Component;

@Component
@ConditionalOnBean(JdbcTemplate.class)
public class OperationTaskRecorder {

    public static final String STATUS_PENDING = "PENDING";
    public static final String STATUS_RUNNING = "RUNNING";
    public static final String STATUS_SUCCESS = "SUCCESS";
    public static final String STATUS_FAILED = "FAILED";
    public static final String STATUS_CANCELLED = "CANCELLED";

    private final JdbcTemplate jdbcTemplate;

    public OperationTaskRecorder(JdbcTemplate jdbcTemplate) {
        this.jdbcTemplate = jdbcTemplate;
    }

    public void pending(HttpServletRequest request, String taskKey, String taskType, String taskName,
                        String targetType, String targetId, String stage, String message) {
        record(request, taskKey, taskType, taskName, targetType, targetId, STATUS_PENDING, 0, stage, message, null);
    }

    public void running(HttpServletRequest request, String taskKey, String taskType, String taskName,
                        String targetType, String targetId, int progress, String stage, String message) {
        record(request, taskKey, taskType, taskName, targetType, targetId, STATUS_RUNNING, progress, stage, message, null);
    }

    public void success(HttpServletRequest request, String taskKey, String taskType, String taskName,
                        String targetType, String targetId, String stage, String message) {
        record(request, taskKey, taskType, taskName, targetType, targetId, STATUS_SUCCESS, 100, stage, message, null);
    }

    public void failed(HttpServletRequest request, String taskKey, String taskType, String taskName,
                       String targetType, String targetId, int progress, String stage, String message, String errorMessage) {
        record(request, taskKey, taskType, taskName, targetType, targetId, STATUS_FAILED, progress, stage, message, errorMessage);
    }

    public void cancelled(HttpServletRequest request, String taskKey, String taskType, String taskName,
                          String targetType, String targetId, String stage, String message) {
        record(request, taskKey, taskType, taskName, targetType, targetId, STATUS_CANCELLED, 0, stage, message, null);
    }

    public void recordResult(HttpServletRequest request, String taskKey, String taskType, String taskName,
                             String targetType, String targetId, String runningStage, Result<?> result) {
        if (isSuccess(result)) {
            success(request, taskKey, taskType, taskName, targetType, targetId, runningStage,
                    result == null ? "任务完成" : result.getMessage());
            return;
        }
        failed(request, taskKey, taskType, taskName, targetType, targetId, 0, runningStage,
                result == null ? "任务失败" : result.getMessage(),
                result == null ? null : result.getMessage());
    }

    public void pending(Integer operatorId, String taskKey, String taskType, String taskName,
                        String targetType, String targetId, String stage, String message) {
        record(operatorId, null, taskKey, taskType, taskName, targetType, targetId,
                STATUS_PENDING, 0, stage, message, null);
    }

    public void running(Integer operatorId, String taskKey, String taskType, String taskName,
                        String targetType, String targetId, int progress, String stage, String message) {
        record(operatorId, null, taskKey, taskType, taskName, targetType, targetId,
                STATUS_RUNNING, progress, stage, message, null);
    }

    public void success(Integer operatorId, String taskKey, String taskType, String taskName,
                        String targetType, String targetId, String stage, String message) {
        record(operatorId, null, taskKey, taskType, taskName, targetType, targetId,
                STATUS_SUCCESS, 100, stage, message, null);
    }

    public void failed(Integer operatorId, String taskKey, String taskType, String taskName,
                       String targetType, String targetId, int progress, String stage, String message, String errorMessage) {
        record(operatorId, null, taskKey, taskType, taskName, targetType, targetId,
                STATUS_FAILED, progress, stage, message, errorMessage);
    }

    public void cancelled(Integer operatorId, String taskKey, String taskType, String taskName,
                          String targetType, String targetId, String stage, String message) {
        record(operatorId, null, taskKey, taskType, taskName, targetType, targetId,
                STATUS_CANCELLED, 0, stage, message, null);
    }

    public void record(HttpServletRequest request, String taskKey, String taskType, String taskName,
                       String targetType, String targetId, String status, int progress,
                       String stage, String message, String errorMessage) {
        record(operatorId(request), header(request, "X-Username"), taskKey, taskType, taskName,
                targetType, targetId, status, progress, stage, message, errorMessage);
    }

    public void record(Integer operatorId, String operatorName, String taskKey, String taskType, String taskName,
                       String targetType, String targetId, String status, int progress,
                       String stage, String message, String errorMessage) {
        validate(taskKey, taskType, taskName, status);
        int normalizedProgress = Math.max(0, Math.min(progress, 100));
        jdbcTemplate.update("""
                        INSERT INTO operation_tasks
                        (task_key, task_type, task_name, target_type, target_id, status, progress,
                         stage, message, error_message, operator_id, operator_name, started_at, finished_at)
                        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
                                CASE WHEN ? = 'RUNNING' THEN NOW() ELSE NULL END,
                                CASE WHEN ? IN ('SUCCESS','FAILED','CANCELLED') THEN NOW() ELSE NULL END)
                        ON DUPLICATE KEY UPDATE
                          task_type = VALUES(task_type),
                          task_name = VALUES(task_name),
                          target_type = VALUES(target_type),
                          target_id = VALUES(target_id),
                          status = VALUES(status),
                          progress = VALUES(progress),
                          stage = VALUES(stage),
                          message = VALUES(message),
                          error_message = VALUES(error_message),
                          operator_id = COALESCE(VALUES(operator_id), operator_id),
                          operator_name = COALESCE(VALUES(operator_name), operator_name),
                          started_at = CASE
                            WHEN VALUES(status) = 'RUNNING' AND started_at IS NULL THEN NOW()
                            ELSE started_at
                          END,
                          finished_at = CASE
                            WHEN VALUES(status) IN ('SUCCESS','FAILED','CANCELLED') THEN NOW()
                            ELSE finished_at
                          END
                        """,
                taskKey,
                taskType,
                limit(taskName, 128),
                targetType,
                targetId,
                status,
                normalizedProgress,
                limit(stage, 128),
                limit(message, 255),
                errorMessage,
                operatorId,
                limit(operatorName, 64),
                status,
                status);
    }

    private boolean isSuccess(Result<?> result) {
        return result != null && result.getCode() != null && result.getCode() == 200;
    }

    private void validate(String taskKey, String taskType, String taskName, String status) {
        if (isBlank(taskKey)) {
            throw new IllegalArgumentException("任务key不能为空");
        }
        if (isBlank(taskType)) {
            throw new IllegalArgumentException("任务类型不能为空");
        }
        if (isBlank(taskName)) {
            throw new IllegalArgumentException("任务名称不能为空");
        }
        if (isBlank(status)) {
            throw new IllegalArgumentException("任务状态不能为空");
        }
    }

    private Integer operatorId(HttpServletRequest request) {
        String adminId = header(request, "X-Admin-Id");
        if (isBlank(adminId)) {
            adminId = header(request, "X-User-Id");
        }
        if (isBlank(adminId)) {
            return null;
        }
        return Integer.parseInt(adminId);
    }

    private String header(HttpServletRequest request, String name) {
        return request == null ? null : request.getHeader(name);
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
