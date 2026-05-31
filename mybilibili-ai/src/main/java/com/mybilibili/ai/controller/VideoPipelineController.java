package com.mybilibili.ai.controller;

import com.mybilibili.ai.audit.AiAuditLogRecorder;
import com.mybilibili.ai.pipeline.PipelineTask;
import com.mybilibili.ai.service.VideoPipelineService;
import com.mybilibili.common.operation.OperationTaskRecorder;
import com.mybilibili.common.vo.Result;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

/**
 * 视频全流程处理控制器
 */
@RestController
@RequestMapping("/api/pipeline")
public class VideoPipelineController {

    @Autowired
    private VideoPipelineService pipelineService;

    @Autowired
    private AiAuditLogRecorder auditLogRecorder;

    @Autowired
    private OperationTaskRecorder operationTaskRecorder;

    /**
     * 提交全流程处理任务
     */
    @PostMapping("/submit")
    public Result<String> submitTask(
            @RequestParam Integer manuscriptId,
            @RequestParam Integer videoId,
            @RequestParam(required = false) Integer uploaderId,
            HttpServletRequest request) {
        operationTaskRecorder.running(request, taskKey("submit", manuscriptId, videoId), "AI_PIPELINE", "全流程任务提交",
                "video", String.valueOf(videoId), 0, "PENDING", "任务提交中");

        boolean success = pipelineService.submitPipelineTask(manuscriptId, videoId, uploaderId);

        Result<String> result;
        if (success) {
            result = Result.success("任务已提交到队列");
            operationTaskRecorder.success(request, taskKey("submit", manuscriptId, videoId), "AI_PIPELINE", "全流程任务提交",
                    "video", String.valueOf(videoId), "QUEUED", result.getMessage());
        } else {
            result = Result.error("任务提交失败，任务可能已存在");
            operationTaskRecorder.failed(request, taskKey("submit", manuscriptId, videoId), "AI_PIPELINE", "全流程任务提交",
                    "video", String.valueOf(videoId), 0, "PENDING", result.getMessage(), result.getMessage());
        }
        auditLogRecorder.record(request, "pipeline_submit", "video", String.valueOf(videoId), result);
        return result;
    }

    /**
     * 取消任务
     */
    @PostMapping("/cancel")
    public Result<String> cancelTask(
            @RequestParam Integer manuscriptId,
            @RequestParam Integer videoId,
            HttpServletRequest request) {

        boolean success = pipelineService.cancelTask(manuscriptId, videoId);

        Result<String> result;
        if (success) {
            result = Result.success("任务已取消");
            operationTaskRecorder.cancelled(request, taskKey("cancel", manuscriptId, videoId), "AI_PIPELINE", "全流程任务取消",
                    "video", String.valueOf(videoId), "CANCELLED", result.getMessage());
        } else {
            result = Result.error("取消失败，任务可能正在处理中或不存在");
            operationTaskRecorder.failed(request, taskKey("cancel", manuscriptId, videoId), "AI_PIPELINE", "全流程任务取消",
                    "video", String.valueOf(videoId), 0, "CANCELLED", result.getMessage(), result.getMessage());
        }
        auditLogRecorder.record(request, "pipeline_cancel", "video", String.valueOf(videoId), result);
        return result;
    }

    /**
     * 重试失败的任务
     */
    @PostMapping("/retry")
    public Result<String> retryTask(
            @RequestParam Integer manuscriptId,
            @RequestParam Integer videoId,
            HttpServletRequest request) {

        boolean success = pipelineService.retryTask(manuscriptId, videoId);

        Result<String> result;
        if (success) {
            result = Result.success("任务已重新提交");
            operationTaskRecorder.running(request, taskKey("retry", manuscriptId, videoId), "AI_PIPELINE", "全流程任务重试",
                    "video", String.valueOf(videoId), 0, "PENDING", result.getMessage());
        } else {
            result = Result.error("重试失败，任务不存在或状态不允许重试");
            operationTaskRecorder.failed(request, taskKey("retry", manuscriptId, videoId), "AI_PIPELINE", "全流程任务重试",
                    "video", String.valueOf(videoId), 0, "PENDING", result.getMessage(), result.getMessage());
        }
        auditLogRecorder.record(request, "pipeline_retry", "video", String.valueOf(videoId), result);
        return result;
    }

    /**
     * 获取任务状态
     */
    @GetMapping("/status")
    public Result<PipelineTask> getTaskStatus(
            @RequestParam Integer manuscriptId,
            @RequestParam Integer videoId) {

        PipelineTask task = pipelineService.getTaskStatus(manuscriptId, videoId);

        if (task != null) {
            return Result.success(task);
        } else {
            return Result.error("任务不存在");
        }
    }

    /**
     * 获取队列统计信息
     */
    @GetMapping("/statistics")
    public Result<Map<String, Object>> getStatistics() {
        return Result.success(pipelineService.getStatistics());
    }

    /**
     * 获取队列中的任务列表
     */
    @GetMapping("/queue")
    public Result<List<PipelineTask>> getQueuedTasks() {
        return Result.success(pipelineService.getQueuedTasks());
    }

    /**
     * 获取正在处理的任务
     */
    @GetMapping("/processing")
    public Result<List<PipelineTask>> getProcessingTasks() {
        return Result.success(pipelineService.getProcessingTasks());
    }

    /**
     * 获取已完成的任务
     */
    @GetMapping("/completed")
    public Result<List<PipelineTask>> getCompletedTasks() {
        return Result.success(pipelineService.getCompletedTasks());
    }

    /**
     * 获取当前正在处理的任务
     */
    @GetMapping("/current")
    public Result<PipelineTask> getCurrentTask() {
        PipelineTask task = pipelineService.getCurrentTask();
        if (task != null) {
            return Result.success(task);
        } else {
            return Result.success(null);
        }
    }

    /**
     * 清空队列（管理员操作）
     */
    @PostMapping("/clear")
    public Result<String> clearQueue(HttpServletRequest request) {
        pipelineService.clearQueue();
        Result<String> result = Result.success("队列已清空");
        operationTaskRecorder.success(request, taskKey("clear", null, null), "AI_PIPELINE", "清空流水线队列",
                "pipeline", "queue", "CLEARED", result.getMessage());
        auditLogRecorder.record(request, "pipeline_clear", "pipeline", "queue", result);
        return result;
    }

    private String taskKey(String action, Integer manuscriptId, Integer videoId) {
        return "pipeline:" + action + ":" + (manuscriptId == null ? "-" : manuscriptId) + ":" + (videoId == null ? "-" : videoId);
    }
}
