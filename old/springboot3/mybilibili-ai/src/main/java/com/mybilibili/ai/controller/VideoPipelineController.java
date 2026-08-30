package com.mybilibili.ai.controller;

import com.mybilibili.ai.pipeline.PipelineTask;
import com.mybilibili.ai.service.VideoPipelineService;
import com.mybilibili.common.vo.Result;
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

    /**
     * 提交全流程处理任务
     */
    @PostMapping("/submit")
    public Result<String> submitTask(
            @RequestParam Integer manuscriptId,
            @RequestParam Integer videoId,
            @RequestParam(required = false) Integer uploaderId) {

        boolean success = pipelineService.submitPipelineTask(manuscriptId, videoId, uploaderId);

        if (success) {
            return Result.success("任务已提交到队列");
        } else {
            return Result.error("任务提交失败，任务可能已存在");
        }
    }

    /**
     * 取消任务
     */
    @PostMapping("/cancel")
    public Result<String> cancelTask(
            @RequestParam Integer manuscriptId,
            @RequestParam Integer videoId) {

        boolean success = pipelineService.cancelTask(manuscriptId, videoId);

        if (success) {
            return Result.success("任务已取消");
        } else {
            return Result.error("取消失败，任务可能正在处理中或不存在");
        }
    }

    /**
     * 重试失败的任务
     */
    @PostMapping("/retry")
    public Result<String> retryTask(
            @RequestParam Integer manuscriptId,
            @RequestParam Integer videoId) {

        boolean success = pipelineService.retryTask(manuscriptId, videoId);

        if (success) {
            return Result.success("任务已重新提交");
        } else {
            return Result.error("重试失败，任务不存在或状态不允许重试");
        }
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
    public Result<String> clearQueue() {
        pipelineService.clearQueue();
        return Result.success("队列已清空");
    }
}
