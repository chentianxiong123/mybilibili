package com.mybilibili.user.controller;

import com.mybilibili.common.vo.Result;
import com.mybilibili.user.entity.OperationTask;
import com.mybilibili.user.service.OperationTaskService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/admin/operation-tasks")
@Tag(name = "任务中心", description = "后台统一任务中心查询")
public class AdminOperationTaskController {

    @Autowired
    private OperationTaskService operationTaskService;

    @GetMapping("/list")
    @Operation(summary = "任务列表", description = "分页查询上传、视频处理、AI流水线、存储迁移等后台任务")
    public Result<Map<String, Object>> getOperationTasks(
            @Parameter(description = "任务类型") @RequestParam(required = false) String taskType,
            @Parameter(description = "任务状态") @RequestParam(required = false) String status,
            @Parameter(description = "对象类型或对象ID") @RequestParam(required = false) String targetKeyword,
            @Parameter(description = "关键词") @RequestParam(required = false) String keyword,
            @Parameter(description = "开始时间") @RequestParam(required = false) String startTime,
            @Parameter(description = "结束时间") @RequestParam(required = false) String endTime,
            @Parameter(description = "页码") @RequestParam(defaultValue = "1") Integer page,
            @Parameter(description = "每页数量") @RequestParam(defaultValue = "10") Integer size) {
        List<OperationTask> tasks = operationTaskService.getTasksByCondition(taskType, status, targetKeyword,
                keyword, startTime, endTime, page, size);
        Integer total = operationTaskService.countByCondition(taskType, status, targetKeyword,
                keyword, startTime, endTime);

        Map<String, Object> data = new HashMap<>();
        data.put("list", tasks);
        data.put("total", total);
        data.put("page", page == null || page < 1 ? 1 : page);
        data.put("size", size == null || size < 1 ? 10 : Math.min(size, 100));
        return Result.success(data);
    }

    @GetMapping("/{id}")
    @Operation(summary = "任务详情")
    public Result<OperationTask> getOperationTask(@PathVariable Long id) {
        OperationTask task = operationTaskService.getById(id);
        return task != null ? Result.success(task) : Result.error("任务不存在");
    }
}
