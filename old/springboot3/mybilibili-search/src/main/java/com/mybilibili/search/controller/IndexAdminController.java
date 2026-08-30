package com.mybilibili.search.controller;

import com.mybilibili.search.document.ManuscriptDocument;
import com.mybilibili.search.service.ManuscriptIndexService;
import com.mybilibili.common.vo.Result;
import com.mybilibili.search.service.VideoSearchService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.elasticsearch.client.elc.ElasticsearchTemplate;
import org.springframework.data.elasticsearch.core.IndexOperations;
import org.springframework.data.elasticsearch.core.query.Query;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.Map;

@RestController
@RequestMapping("/search/admin/index")
@Tag(name = "索引管理接口", description = "Elasticsearch 索引管理相关操作")
public class IndexAdminController {

    @Autowired
    private VideoSearchService videoSearchService;

    @Autowired
    private ElasticsearchTemplate elasticsearchTemplate;

    @Autowired
    private ManuscriptIndexService manuscriptIndexService;

    @GetMapping("/status")
    @Operation(summary = "获取索引状态", description = "获取当前索引的文档数量、索引名称、健康状态")
    public Result<Map<String, Object>> getIndexStatus() {
        Map<String, Object> result = new HashMap<>();

        try {
            IndexOperations indexOps = elasticsearchTemplate.indexOps(ManuscriptDocument.class);

            Map<String, Object> indexInfo = new HashMap<>();
            indexInfo.put("indexName", "manuscripts");

            if (!indexOps.exists()) {
                indexInfo.put("status", "not_found");
                indexInfo.put("indexedCount", 0);
                indexInfo.put("message", "索引不存在，请先执行批量索引或重建索引");
                return Result.success("索引不存在", indexInfo);
            }

            indexInfo.put("status", "active");
            Query query = Query.findAll();
            long count = elasticsearchTemplate.count(query, ManuscriptDocument.class);
            indexInfo.put("indexedCount", count);

            result = indexInfo;

            return Result.success("获取成功", result);
        } catch (Exception e) {
            return Result.error("获取索引状态失败：" + e.getMessage());
        }
    }

    @PostMapping("/bulk")
    @Operation(summary = "批量索引稿件", description = "将所有已上架稿件批量索引到 Elasticsearch")
    public Result<Map<String, Object>> bulkIndex() {
        Map<String, Object> result = new HashMap<>();

        try {
            int count = manuscriptIndexService.bulkIndex();

            result.put("status", "success");
            result.put("message", "批量索引完成");
            result.put("indexedCount", count);

            return Result.success("批量索引成功", result);
        } catch (Exception e) {
            return Result.error("批量索引失败：" + e.getMessage());
        }
    }

    @PostMapping("/rebuild")
    @Operation(summary = "重建索引", description = "删除旧索引并重新创建，然后批量导入数据")
    public Result<Map<String, Object>> rebuildIndex() {
        Map<String, Object> result = new HashMap<>();

        try {
            int count = manuscriptIndexService.rebuildIndex();

            result.put("status", "success");
            result.put("message", "索引重建完成");
            result.put("indexedCount", count);

            return Result.success("重建索引成功", result);
        } catch (Exception e) {
            return Result.error("重建索引失败：" + e.getMessage());
        }
    }

    @PostMapping("/refresh")
    @Operation(summary = "刷新索引", description = "刷新索引使最近的更改立即可搜索")
    public Result<Map<String, Object>> refreshIndex() {
        Map<String, Object> result = new HashMap<>();

        try {
            IndexOperations indexOps = elasticsearchTemplate.indexOps(ManuscriptDocument.class);
            indexOps.refresh();

            result.put("status", "success");
            result.put("message", "索引刷新成功");

            return Result.success("刷新索引成功", result);
        } catch (Exception e) {
            return Result.error("刷新索引失败：" + e.getMessage());
        }
    }

    @PostMapping("/incremental")
    @Operation(summary = "增量索引", description = "索引所有已上架但尚未被索引的稿件")
    public Result<Map<String, Object>> incrementalIndex() {
        Map<String, Object> result = new HashMap<>();

        try {
            int count = manuscriptIndexService.incrementalIndex();

            result.put("status", "success");
            result.put("message", count > 0
                    ? "增量索引完成，新增 " + count + " 条稿件"
                    : "增量索引完成，没有缺失的稿件");
            result.put("indexedCount", count);

            return Result.success("增量索引成功", result);
        } catch (Exception e) {
            return Result.error("增量索引失败：" + e.getMessage());
        }
    }
}
