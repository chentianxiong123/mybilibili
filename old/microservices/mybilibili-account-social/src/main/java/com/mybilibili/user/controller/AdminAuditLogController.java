package com.mybilibili.user.controller;

import com.mybilibili.common.vo.Result;
import com.mybilibili.user.entity.AuditLog;
import com.mybilibili.user.service.AuditLogService;
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
@RequestMapping("/admin/audit-logs")
@Tag(name = "审计日志管理", description = "管理员操作审计日志查询")
public class AdminAuditLogController {

    @Autowired
    private AuditLogService auditLogService;

    @GetMapping("/list")
    @Operation(summary = "审计日志列表", description = "分页查询后台敏感操作审计日志")
    public Result<Map<String, Object>> getAuditLogs(
            @Parameter(description = "操作者ID或用户名") @RequestParam(required = false) String operatorKeyword,
            @Parameter(description = "模块") @RequestParam(required = false) String module,
            @Parameter(description = "动作") @RequestParam(required = false) String action,
            @Parameter(description = "结果：1成功 0失败") @RequestParam(required = false) Integer result,
            @Parameter(description = "对象类型或对象ID") @RequestParam(required = false) String targetKeyword,
            @Parameter(description = "开始时间") @RequestParam(required = false) String startTime,
            @Parameter(description = "结束时间") @RequestParam(required = false) String endTime,
            @Parameter(description = "页码") @RequestParam(defaultValue = "1") Integer page,
            @Parameter(description = "每页数量") @RequestParam(defaultValue = "10") Integer size) {
        List<AuditLog> logs = auditLogService.getLogsByCondition(operatorKeyword, module, action, result,
                targetKeyword, startTime, endTime, page, size);
        Integer total = auditLogService.countByCondition(operatorKeyword, module, action, result,
                targetKeyword, startTime, endTime);

        Map<String, Object> data = new HashMap<>();
        data.put("list", logs);
        data.put("total", total);
        data.put("page", page == null || page < 1 ? 1 : page);
        data.put("size", size == null || size < 1 ? 10 : Math.min(size, 100));
        return Result.success(data);
    }

    @GetMapping("/{id}")
    @Operation(summary = "审计日志详情")
    public Result<AuditLog> getAuditLog(@PathVariable Long id) {
        AuditLog auditLog = auditLogService.getById(id);
        return auditLog != null ? Result.success(auditLog) : Result.error("审计日志不存在");
    }
}
