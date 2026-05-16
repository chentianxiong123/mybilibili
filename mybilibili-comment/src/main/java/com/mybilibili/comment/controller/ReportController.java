package com.mybilibili.comment.controller;

import com.mybilibili.comment.service.ReportService;
import com.mybilibili.common.vo.Result;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.security.SecurityRequirement;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import javax.servlet.http.HttpServletRequest;
import java.util.Map;

@RestController
@RequestMapping("/report")
@Tag(name = "举报接口", description = "内容举报相关操作")
public class ReportController {

    @Autowired
    private ReportService reportService;

    @PostMapping("/submit")
    @Operation(summary = "提交举报", description = "用户提交内容举报")
    @SecurityRequirement(name = "JWT")
    public Result<?> submitReport(@RequestBody Map<String, Object> request,
                                   HttpServletRequest httpRequest) {
        Integer userId = getUserIdFromRequest(httpRequest);
        if (userId == null) {
            return Result.error("用户未登录");
        }

        String targetType = (String) request.get("targetType");
        Integer targetId = (Integer) request.get("targetId");
        Integer manuscriptId = (Integer) request.get("manuscriptId");
        String reason = (String) request.get("reason");
        String description = (String) request.get("description");

        if (targetType == null || targetId == null || reason == null || reason.isEmpty()) {
            return Result.error("缺少必要参数");
        }

        return reportService.submitReport(userId, targetType, targetId, manuscriptId, reason, description);
    }

    private Integer getUserIdFromRequest(HttpServletRequest request) {
        String userIdStr = request.getHeader("X-User-Id");
        if (userIdStr != null && !userIdStr.isEmpty()) {
            try {
                return Integer.parseInt(userIdStr);
            } catch (NumberFormatException e) {
                return null;
            }
        }
        return null;
    }
}
