package com.mybilibili.ai.controller;

import com.mybilibili.ai.config.DynamicModerationClient;
import com.mybilibili.ai.entity.AiApiConfig;
import com.mybilibili.ai.service.AiApiConfigService;
import com.mybilibili.ai.service.ModerationProvider;
import com.mybilibili.ai.service.ModerationProvider.ModerateRequest;
import com.mybilibili.ai.service.ModerationProvider.ModerationResult;
import com.mybilibili.ai.util.AiUsageLogger;
import com.mybilibili.common.vo.Result;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

/**
 * 内容审核 AI 接口。
 * 供 comment 服务调用，进行评论/回复预过滤和举报内容深度审核。
 */
@RestController
@RequestMapping("/ai/review")
@Tag(name = "内容审核AI接口", description = "评论预过滤和举报内容深度审核")
public class ContentReviewAiController {

    @Autowired
    private DynamicModerationClient moderationClient;

    @Autowired
    private AiApiConfigService aiApiConfigService;

    @Autowired
    private AiUsageLogger aiUsageLogger;

    @PostMapping("/content")
    @Operation(summary = "内容安全审核", description = "检测文本是否违规，返回违规类型和原因")
    public Result<List<String>> moderateContent(
            @Parameter(description = "待审核文本") @RequestParam String content,
            @Parameter(description = "审核场景：COMMENT=评论预过滤，REPLY=回复预过滤，REPORT=举报审核") @RequestParam(required = false, defaultValue = "COMMENT") String scene) {

        if (content == null || content.isEmpty()) {
            return Result.success(List.of());
        }

        ModerationProvider provider = moderationClient.getProvider();
        if (provider == null || !provider.isAvailable()) {
            return Result.success(List.of());
        }

        ModerateRequest request = ModerateRequest.of(content, scene);
        ModerationProvider.ModerationResult result = (ModerationProvider.ModerationResult) provider.invoke(request);

        if (result.isPassed()) {
            return Result.success(List.of());
        } else {
            List<String> reasons = result.getViolationTypes();
            if (reasons == null || reasons.isEmpty()) {
                reasons = List.of(result.getReason() != null ? result.getReason() : "内容违规");
            }
            return Result.success(reasons);
        }
    }

    @PostMapping("/comment")
    @Operation(summary = "评论预过滤", description = "发评论前使用 REVIEW 渠道过滤，拦截 spam/广告/低质量")
    public Result<Map<String, Object>> moderateComment(@RequestParam String content) {
        ModerationProvider provider = moderationClient.getProvider();
        String model = reviewModel();
        if (provider == null || !provider.isAvailable()) {
            aiUsageLogger.log("REVIEW", model, null, null, 0, false, "provider unavailable");
            return Result.success(Map.of("passed", true, "reason", "审核服务不可用"));
        }

        long start = System.currentTimeMillis();
        try {
            ModerateRequest request = ModerateRequest.of(content, "COMMENT");
            ModerationProvider.ModerationResult result = (ModerationProvider.ModerationResult) provider.invoke(request);

            aiUsageLogger.log("REVIEW", model, null, null, System.currentTimeMillis() - start, true, null);
            return Result.success(Map.of(
                    "passed", result.isPassed(),
                    "reason", result.getReason() != null ? result.getReason() : ""
            ));
        } catch (Exception e) {
            aiUsageLogger.log("REVIEW", model, null, null, System.currentTimeMillis() - start, false, e.getMessage());
            throw e;
        }
    }

    @PostMapping("/reply")
    @Operation(summary = "回复预过滤", description = "发回复前使用 REVIEW 渠道过滤")
    public Result<Map<String, Object>> moderateReply(@RequestParam String content) {
        ModerationProvider provider = moderationClient.getProvider();
        String model = reviewModel();
        if (provider == null || !provider.isAvailable()) {
            aiUsageLogger.log("REVIEW", model, null, null, 0, false, "provider unavailable");
            return Result.success(Map.of("passed", true, "reason", "审核服务不可用"));
        }

        long start = System.currentTimeMillis();
        try {
            ModerateRequest request = ModerateRequest.of(content, "REPLY");
            ModerationProvider.ModerationResult result = (ModerationProvider.ModerationResult) provider.invoke(request);

            aiUsageLogger.log("REVIEW", model, null, null, System.currentTimeMillis() - start, true, null);
            return Result.success(Map.of(
                    "passed", result.isPassed(),
                    "reason", result.getReason() != null ? result.getReason() : ""
            ));
        } catch (Exception e) {
            aiUsageLogger.log("REVIEW", model, null, null, System.currentTimeMillis() - start, false, e.getMessage());
            throw e;
        }
    }

    @PostMapping("/report")
    @Operation(summary = "举报内容审核", description = "举报内容使用 REVIEW 渠道深度审核，识别辱骂/政治/色情等")
    public Result<Map<String, Object>> moderateReport(@RequestParam String content) {
        ModerationProvider provider = moderationClient.getProvider();
        String model = reviewModel();
        if (provider == null || !provider.isAvailable()) {
            aiUsageLogger.log("REVIEW", model, null, null, 0, false, "provider unavailable");
            return Result.success(Map.of("passed", true, "reason", "审核服务不可用"));
        }

        long start = System.currentTimeMillis();
        try {
            ModerateRequest request = ModerateRequest.of(content, "REPORT");
            ModerationProvider.ModerationResult result = (ModerationProvider.ModerationResult) provider.invoke(request);

            aiUsageLogger.log("REVIEW", model, null, null, System.currentTimeMillis() - start, true, null);
            return Result.success(Map.of(
                    "passed", result.isPassed(),
                    "reason", result.getReason() != null ? result.getReason() : "",
                    "violationTypes", result.getViolationTypes() != null ? result.getViolationTypes() : List.of()
            ));
        } catch (Exception e) {
            aiUsageLogger.log("REVIEW", model, null, null, System.currentTimeMillis() - start, false, e.getMessage());
            throw e;
        }
    }

    private String reviewModel() {
        AiApiConfig config = aiApiConfigService.getConfigForFeature("REVIEW");
        return config != null ? config.getModel() : null;
    }
}
