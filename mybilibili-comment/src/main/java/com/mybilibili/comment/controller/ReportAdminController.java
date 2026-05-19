package com.mybilibili.comment.controller;

import com.mybilibili.comment.service.ReportService;
import com.mybilibili.common.vo.Result;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import javax.servlet.http.HttpServletRequest;
import java.util.Map;

@RestController
@RequestMapping("/admin/report")
@Tag(name = "举报管理接口", description = "管理员处理举报")
public class ReportAdminController {

    @Autowired
    private ReportService reportService;

    @GetMapping("/list")
    @Operation(summary = "获取举报列表")
    public Result<?> getReportList(
            @Parameter(description = "状态: PENDING/RESOLVED/REJECTED") @RequestParam(required = false) String status,
            @Parameter(description = "目标类型: COMMENT/REPLY/DYNAMIC_COMMENT") @RequestParam(required = false) String targetType,
            @Parameter(description = "页码") @RequestParam(defaultValue = "1") Integer page,
            @Parameter(description = "每页数量") @RequestParam(defaultValue = "10") Integer size) {
        return reportService.getReportList(status, targetType, page, size);
    }

    @PutMapping("/process/{id}")
    @Operation(summary = "处理举报")
    public Result<?> processReport(
            @PathVariable Integer id,
            @RequestBody Map<String, Object> request,
            HttpServletRequest httpRequest) {
        String action = (String) request.get("action"); // resolve 或 reject
        String adminRemark = (String) request.get("adminRemark");

        Integer adminId = null;
        String adminIdStr = httpRequest.getHeader("X-User-Id");
        if (adminIdStr != null) {
            try {
                adminId = Integer.parseInt(adminIdStr);
            } catch (NumberFormatException ignored) {}
        }

        return reportService.processReport(id, action, adminRemark, adminId);
    }

    @PutMapping("/ai-review-result")
    @Operation(summary = "AI审核结果回调", hidden = true)
    public Result<?> updateAiReviewResult(@RequestBody Map<String, Object> result) {
        return reportService.updateAiReviewResult(result);
    }
}
