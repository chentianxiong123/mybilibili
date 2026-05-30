package com.mybilibili.user.controller;

import com.mybilibili.common.vo.Result;
import com.mybilibili.user.entity.LoginLog;
import com.mybilibili.user.service.LoginLogService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/admin/login-logs")
@Tag(name = "登录日志管理", description = "管理员查看登录日志")
public class AdminLoginLogController {

    @Autowired
    private LoginLogService loginLogService;

    @GetMapping("/list")
    @Operation(summary = "登录日志列表", description = "分页查询登录日志，支持IP、用户ID、时间范围筛选")
    public Result<Map<String, Object>> getLoginLogs(
            @Parameter(description = "IP模糊搜索") @RequestParam(required = false) String ip,
            @Parameter(description = "用户ID") @RequestParam(required = false) Integer userId,
            @Parameter(description = "状态：1成功 0失败") @RequestParam(required = false) Integer status,
            @Parameter(description = "开始时间") @RequestParam(required = false) String startTime,
            @Parameter(description = "结束时间") @RequestParam(required = false) String endTime,
            @Parameter(description = "页码") @RequestParam(defaultValue = "1") Integer page,
            @Parameter(description = "每页数量") @RequestParam(defaultValue = "10") Integer size) {
        if (page == null || page < 1) page = 1;
        if (size == null || size < 1) size = 10;
        Integer offset = (page - 1) * size;
        List<LoginLog> logs = loginLogService.getLogsByCondition(ip, userId, status, startTime, endTime, page, size);
        Integer total = loginLogService.countByCondition(ip, userId, status, startTime, endTime);
        Map<String, Object> data = new HashMap<>();
        data.put("list", logs);
        data.put("total", total);
        data.put("page", page);
        data.put("size", size);
        return Result.success(data);
    }

    @GetMapping("/user/{userId}")
    @Operation(summary = "指定用户的登录日志")
    public Result<Map<String, Object>> getUserLoginLogs(
            @PathVariable Integer userId,
            @RequestParam(defaultValue = "1") Integer page,
            @RequestParam(defaultValue = "10") Integer size) {
        List<LoginLog> logs = loginLogService.getUserLogs(userId, page, size);
        Integer total = loginLogService.countUserLogs(userId);
        Map<String, Object> data = new HashMap<>();
        data.put("list", logs);
        data.put("total", total);
        data.put("page", page);
        data.put("size", size);
        return Result.success(data);
    }
}