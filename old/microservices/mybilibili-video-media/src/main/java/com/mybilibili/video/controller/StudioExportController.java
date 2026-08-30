package com.mybilibili.video.controller;

import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.vo.Result;
import com.mybilibili.video.studio.StudioExportTaskRequest;
import com.mybilibili.video.studio.StudioExportTaskResponse;
import com.mybilibili.video.studio.StudioExportTaskService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/studio/export-tasks")
@Tag(name = "剪辑工作室导出", description = "云端导出任务接口")
public class StudioExportController {
    private final StudioExportTaskService taskService;

    public StudioExportController(StudioExportTaskService taskService) {
        this.taskService = taskService;
    }

    @PostMapping
    @Operation(summary = "创建云端导出任务")
    public Result<StudioExportTaskResponse> createTask(@RequestBody StudioExportTaskRequest body,
                                                       HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            return Result.success("云端导出任务已创建", taskService.createTask(userId, body));
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/{taskId}")
    @Operation(summary = "查询云端导出任务")
    public Result<StudioExportTaskResponse> getTask(@PathVariable String taskId,
                                                    HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            return Result.success("获取成功", taskService.getTask(userId, taskId));
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping("/{taskId}/cancel")
    @Operation(summary = "取消云端导出任务")
    public Result<StudioExportTaskResponse> cancelTask(@PathVariable String taskId,
                                                       HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            return Result.success("已取消", taskService.cancelTask(userId, taskId));
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}
