package com.mybilibili.admin.controller;

import com.mybilibili.admin.service.ManuscriptIndexService;
import com.mybilibili.common.vo.Result;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

@RestController
@RequestMapping("/api/search/admin")
@Tag(name = "索引管理", description = "Elasticsearch索引管理接口")
public class IndexManagerController {

    @Autowired
    private ManuscriptIndexService manuscriptIndexService;

    /**
     * 获取索引状态
     *
     * @return 索引状态信息
     */
    @GetMapping("/index/status")
    @Operation(summary = "获取索引状态", description = "获取Elasticsearch索引的状态信息")
    public Result<Map<String, Object>> getIndexStatus() {
        try {
            Map<String, Object> status = manuscriptIndexService.getIndexStatus();
            return Result.success("获取成功", status);
        } catch (Exception e) {
            return Result.error("获取索引状态失败: " + e.getMessage());
        }
    }

    /**
     * 批量索引所有已上架稿件
     *
     * @return 操作结果
     */
    @PostMapping("/index/bulk")
    @Operation(summary = "批量索引稿件", description = "批量导入所有已上架稿件到Elasticsearch")
    public Result<Map<String, Object>> bulkIndexAll() {
        try {
            Map<String, Object> result = manuscriptIndexService.bulkIndexAllPublished();
            if ((Boolean) result.get("success")) {
                return Result.success((String) result.get("message"), result);
            } else {
                return Result.error((String) result.get("message"));
            }
        } catch (Exception e) {
            return Result.error("批量索引失败: " + e.getMessage());
        }
    }

    /**
     * 重建索引
     *
     * @return 操作结果
     */
    @PostMapping("/index/rebuild")
    @Operation(summary = "重建索引", description = "清空所有索引并重新导入所有已上架稿件")
    public Result<Map<String, Object>> rebuildIndex() {
        try {
            Map<String, Object> result = manuscriptIndexService.rebuildIndex();
            if ((Boolean) result.get("success")) {
                return Result.success((String) result.get("message"), result);
            } else {
                return Result.error((String) result.get("message"));
            }
        } catch (Exception e) {
            return Result.error("重建索引失败: " + e.getMessage());
        }
    }

    /**
     * 增量索引
     *
     * @param minutes 最近多少分钟内上架的稿件
     * @return 操作结果
     */
    @PostMapping("/index/incremental")
    @Operation(summary = "增量索引", description = "索引最近上架的稿件")
    public Result<Map<String, Object>> incrementalIndex(
            @Parameter(description = "最近多少分钟") @RequestParam(value = "minutes", defaultValue = "60") int minutes) {
        try {
            Map<String, Object> result = manuscriptIndexService.incrementalIndex(minutes);
            if ((Boolean) result.get("success")) {
                return Result.success((String) result.get("message"), result);
            } else {
                return Result.error((String) result.get("message"));
            }
        } catch (Exception e) {
            return Result.error("增量索引失败: " + e.getMessage());
        }
    }
}
