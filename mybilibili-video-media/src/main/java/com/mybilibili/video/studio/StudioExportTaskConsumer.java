package com.mybilibili.video.studio;

import com.mybilibili.mq.MQConstants;
import com.mybilibili.mq.StudioExportTaskMessage;
import lombok.extern.slf4j.Slf4j;
import org.apache.rocketmq.spring.annotation.ConsumeMode;
import org.apache.rocketmq.spring.annotation.RocketMQMessageListener;
import org.apache.rocketmq.spring.core.RocketMQListener;
import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;

@Component
@Slf4j
@RocketMQMessageListener(
        topic = MQConstants.TOPIC_STUDIO_EXPORT,
        consumerGroup = MQConstants.GROUP_STUDIO_EXPORT,
        consumeMode = ConsumeMode.ORDERLY
)
public class StudioExportTaskConsumer implements RocketMQListener<StudioExportTaskMessage> {

    private final StudioExportTaskService taskService;

    public StudioExportTaskConsumer(StudioExportTaskService taskService) {
        this.taskService = taskService;
    }

    @Override
    public void onMessage(StudioExportTaskMessage message) {
        if (message == null || !StringUtils.hasText(message.getTaskId())) {
            log.warn("剪辑云端导出消息无效: {}", message);
            return;
        }

        String taskId = message.getTaskId();
        try {
            taskService.markRunning(taskId, 10, "RENDERER_RECEIVED", "渲染节点已收到任务");
            taskService.getWorkerTask(taskId);
            taskService.markFailed(taskId, "RENDERER_NOT_READY", "云端 FFmpeg 渲染器尚未接入，当前只完成任务编排链路");
        } catch (Exception e) {
            log.warn("剪辑云端导出任务处理失败: taskId={}, error={}", taskId, e.getMessage());
            try {
                taskService.markFailed(taskId, "WORKER_FAILED", e.getMessage());
            } catch (Exception ignored) {
                log.warn("剪辑云端导出任务失败状态写入失败: taskId={}", taskId);
            }
        }
    }
}
