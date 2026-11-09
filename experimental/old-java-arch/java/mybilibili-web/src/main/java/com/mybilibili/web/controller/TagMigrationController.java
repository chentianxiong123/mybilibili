package com.mybilibili.web.controller;

import com.mybilibili.common.vo.Result;
import com.mybilibili.web.service.TagMigrationService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Slf4j
@RestController
@RequestMapping("/admin/tag-migration")
@Tag(name = "标签迁移管理", description = "将MySQL标签数据迁移到Redis的管理接口")
public class TagMigrationController {

    @Autowired
    private TagMigrationService tagMigrationService;

    @PostMapping("/migrate-all")
    @Operation(summary = "迁移所有视频标签", description = "将所有视频的标签从MySQL迁移到Redis")
    public Result<Map<String, Object>> migrateAll() {
        try {
            int count = tagMigrationService.migrateAllVideoTags();
            Map<String, Object> result = new HashMap<>();
            result.put("migratedCount", count);
            result.put("message", "迁移完成");
            return Result.success("迁移成功", result);
        } catch (Exception e) {
            log.error("标签迁移失败", e);
            return Result.error("迁移失败: " + e.getMessage());
        }
    }

    @PostMapping("/migrate-batch")
    @Operation(summary = "批量迁移视频标签", description = "批量迁移指定视频的标签")
    public Result<Map<String, Object>> migrateBatch(@RequestBody List<Integer> videoIds) {
        try {
            int count = tagMigrationService.migrateVideoTagsBatch(videoIds);
            Map<String, Object> result = new HashMap<>();
            result.put("migratedCount", count);
            result.put("totalCount", videoIds.size());
            result.put("message", "批量迁移完成");
            return Result.success("批量迁移成功", result);
        } catch (Exception e) {
            log.error("批量标签迁移失败", e);
            return Result.error("批量迁移失败: " + e.getMessage());
        }
    }

    @PostMapping("/migrate/{videoId}")
    @Operation(summary = "迁移单个视频标签", description = "迁移指定视频的标签")
    public Result<Map<String, Object>> migrateSingle(@PathVariable Integer videoId) {
        try {
            boolean success = tagMigrationService.migrateVideoTags(videoId);
            Map<String, Object> result = new HashMap<>();
            result.put("videoId", videoId);
            result.put("success", success);
            return Result.success(success ? "迁移成功" : "迁移失败", result);
        } catch (Exception e) {
            log.error("视频 {} 标签迁移失败", videoId, e);
            return Result.error("迁移失败: " + e.getMessage());
        }
    }

    @GetMapping("/verify/{videoId}")
    @Operation(summary = "验证迁移结果", description = "验证指定视频的标签迁移结果")
    public Result<TagMigrationService.TagMigrationResult> verify(@PathVariable Integer videoId) {
        try {
            TagMigrationService.TagMigrationResult result = tagMigrationService.verifyMigration(videoId);
            return Result.success("验证完成", result);
        } catch (Exception e) {
            log.error("验证视频 {} 标签迁移失败", videoId, e);
            return Result.error("验证失败: " + e.getMessage());
        }
    }

    @PostMapping("/clear-redis")
    @Operation(summary = "清空Redis标签数据", description = "清空Redis中的所有标签数据（谨慎使用）")
    public Result<Map<String, Object>> clearRedis() {
        try {
            boolean success = tagMigrationService.clearAllRedisTags();
            Map<String, Object> result = new HashMap<>();
            result.put("success", success);
            result.put("message", success ? "清空完成" : "清空失败");
            return Result.success(success ? "清空成功" : "清空失败", result);
        } catch (Exception e) {
            log.error("清空Redis标签数据失败", e);
            return Result.error("清空失败: " + e.getMessage());
        }
    }
}
