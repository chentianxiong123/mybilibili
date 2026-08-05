package com.mybilibili.video.studio;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.mybilibili.common.exception.BusinessException;
import com.mybilibili.mq.MQConstants;
import com.mybilibili.mq.StudioExportTaskMessage;
import lombok.extern.slf4j.Slf4j;
import org.apache.rocketmq.spring.core.RocketMQTemplate;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;
import org.springframework.util.StringUtils;

import java.time.Duration;
import java.util.UUID;

@Service
@Slf4j
public class StudioExportTaskService {
    private static final String KEY_PREFIX = "studio:export:task:";
    private static final Duration TASK_TTL = Duration.ofDays(7);

    private final StringRedisTemplate redisTemplate;
    private final ObjectMapper objectMapper;
    private final RocketMQTemplate rocketMQTemplate;

    public StudioExportTaskService(StringRedisTemplate redisTemplate,
                                   ObjectMapper objectMapper,
                                   RocketMQTemplate rocketMQTemplate) {
        this.redisTemplate = redisTemplate;
        this.objectMapper = objectMapper;
        this.rocketMQTemplate = rocketMQTemplate;
    }

    public StudioExportTaskResponse createTask(Integer userId, StudioExportTaskRequest request) {
        if (userId == null) {
            throw new BusinessException("用户未登录");
        }
        if (request == null || request.getProject() == null) {
            throw new BusinessException("项目数据不能为空");
        }

        String taskId = UUID.randomUUID().toString().replace("-", "");
        long now = System.currentTimeMillis();

        StudioExportTaskRecord record = new StudioExportTaskRecord();
        record.setTaskId(taskId);
        record.setUserId(userId);
        record.setProjectId(StringUtils.hasText(request.getProjectId()) ? request.getProjectId().trim() : null);
        record.setProjectName(StringUtils.hasText(request.getProjectName()) ? request.getProjectName().trim() : "未命名项目");
        record.setStatus(StudioExportStatus.PENDING);
        record.setProgress(0);
        record.setStage("QUEUED");
        record.setMessage("云端导出任务已创建，等待渲染节点处理");
        record.setProjectJson(toJson(request.getProject()));
        record.setExportSettingsJson(toJson(request.getExportSettings()));
        record.setCreatedAt(now);
        record.setUpdatedAt(now);

        save(record);
        publishTask(record);
        return StudioExportTaskResponse.fromRecord(load(taskId));
    }

    public StudioExportTaskResponse getTask(Integer userId, String taskId) {
        StudioExportTaskRecord record = load(taskId);
        ensureOwner(record, userId);
        return StudioExportTaskResponse.fromRecord(record);
    }

    public StudioExportTaskResponse cancelTask(Integer userId, String taskId) {
        StudioExportTaskRecord record = load(taskId);
        ensureOwner(record, userId);
        if (StudioExportStatus.SUCCEEDED.equals(record.getStatus()) || StudioExportStatus.FAILED.equals(record.getStatus())) {
            return StudioExportTaskResponse.fromRecord(record);
        }
        record.setStatus(StudioExportStatus.CANCELLED);
        record.setProgress(Math.max(record.getProgress() == null ? 0 : record.getProgress(), 0));
        record.setStage("CANCELLED");
        record.setMessage("云端导出任务已取消");
        record.setUpdatedAt(System.currentTimeMillis());
        save(record);
        return StudioExportTaskResponse.fromRecord(record);
    }

    public StudioExportTaskRecord getWorkerTask(String taskId) {
        return load(taskId);
    }

    public void markQueued(String taskId) {
        updateForWorker(taskId, StudioExportStatus.PENDING, 0, "QUEUED", "云端导出任务已入队，等待渲染节点处理", null, null);
    }

    public void markRunning(String taskId, Integer progress, String stage, String message) {
        updateForWorker(taskId, StudioExportStatus.RUNNING, progress, stage, message, null, null);
    }

    public void markSucceeded(String taskId, String outputUrl) {
        updateForWorker(taskId, StudioExportStatus.SUCCEEDED, 100, "SUCCEEDED", "云端导出完成", null, outputUrl);
    }

    public void markFailed(String taskId, String stage, String errorMessage) {
        updateForWorker(taskId, StudioExportStatus.FAILED, null, stage, "云端导出失败", errorMessage, null);
    }

    private StudioExportTaskRecord load(String taskId) {
        if (!StringUtils.hasText(taskId) || !taskId.matches("^[a-f0-9]{32}$")) {
            throw new BusinessException("导出任务不存在");
        }
        String value = redisTemplate.opsForValue().get(key(taskId));
        if (!StringUtils.hasText(value)) {
            throw new BusinessException("导出任务不存在或已过期");
        }
        try {
            return objectMapper.readValue(value, StudioExportTaskRecord.class);
        } catch (Exception e) {
            throw new BusinessException("导出任务状态读取失败");
        }
    }

    private void save(StudioExportTaskRecord record) {
        try {
            redisTemplate.opsForValue().set(key(record.getTaskId()), objectMapper.writeValueAsString(record), TASK_TTL);
        } catch (Exception e) {
            throw new BusinessException("导出任务保存失败");
        }
    }

    private void ensureOwner(StudioExportTaskRecord record, Integer userId) {
        if (userId == null || record.getUserId() == null || !record.getUserId().equals(userId)) {
            throw new BusinessException("无权访问该导出任务");
        }
    }

    private String toJson(Object value) {
        if (value == null) {
            return "{}";
        }
        try {
            return objectMapper.writeValueAsString(value);
        } catch (Exception e) {
            throw new BusinessException("导出任务数据序列化失败");
        }
    }

    private String key(String taskId) {
        return KEY_PREFIX + taskId;
    }

    private void updateForWorker(String taskId,
                                 String status,
                                 Integer progress,
                                 String stage,
                                 String message,
                                 String errorMessage,
                                 String outputUrl) {
        StudioExportTaskRecord record = load(taskId);
        if (isTerminal(record.getStatus())) {
            return;
        }
        record.setStatus(status);
        if (progress != null) {
            record.setProgress(Math.max(0, Math.min(100, progress)));
        }
        if (StringUtils.hasText(stage)) {
            record.setStage(stage);
        }
        if (StringUtils.hasText(message)) {
            record.setMessage(message);
        }
        record.setErrorMessage(errorMessage);
        if (StringUtils.hasText(outputUrl)) {
            record.setOutputUrl(outputUrl);
        }
        record.setUpdatedAt(System.currentTimeMillis());
        save(record);
    }

    private boolean isTerminal(String status) {
        return StudioExportStatus.SUCCEEDED.equals(status)
                || StudioExportStatus.FAILED.equals(status)
                || StudioExportStatus.CANCELLED.equals(status);
    }

    private void publishTask(StudioExportTaskRecord record) {
        try {
            rocketMQTemplate.asyncSend(
                    MQConstants.TOPIC_STUDIO_EXPORT,
                    new StudioExportTaskMessage(
                            record.getTaskId(),
                            record.getUserId(),
                            record.getProjectId(),
                            record.getCreatedAt()
                    ),
                    new org.apache.rocketmq.client.producer.SendCallback() {
                        @Override
                        public void onSuccess(org.apache.rocketmq.client.producer.SendResult sendResult) {
                            markQueued(record.getTaskId());
                            log.debug("云端导出任务已入队: taskId={}, msgId={}", record.getTaskId(), sendResult.getMsgId());
                        }

                        @Override
                        public void onException(Throwable e) {
                            markFailed(record.getTaskId(), "QUEUE_FAILED", "任务入队失败：" + e.getMessage());
                            log.warn("云端导出任务入队失败: taskId={}, error={}", record.getTaskId(), e.getMessage());
                        }
                    }
            );
        } catch (Exception e) {
            markFailed(record.getTaskId(), "QUEUE_FAILED", "任务入队异常：" + e.getMessage());
            log.warn("云端导出任务入队异常: taskId={}, error={}", record.getTaskId(), e.getMessage());
        }
    }
}
