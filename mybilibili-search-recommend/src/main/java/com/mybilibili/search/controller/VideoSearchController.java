package com.mybilibili.search.controller;

import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.VideoSearchVO;
import com.mybilibili.search.service.VideoSearchService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.domain.Page;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/search")
@Tag(name = "视频搜索", description = "视频搜索相关接口")
public class VideoSearchController {

    @Autowired
    private VideoSearchService videoSearchService;

    @GetMapping("/videos")
    @Operation(summary = "搜索视频", description = "根据关键词搜索视频，支持分类、标签、UP主过滤和多种排序方式")
    public Result<Map<String, Object>> searchVideos(
            @Parameter(description = "搜索关键词") @RequestParam(value = "keyword", required = false) String keyword,
            @Parameter(description = "分类ID") @RequestParam(value = "categoryId", required = false) Integer categoryId,
            @Parameter(description = "标签") @RequestParam(value = "tag", required = false) String tag,
            @Parameter(description = "UP主ID") @RequestParam(value = "userId", required = false) Integer userId,
            @Parameter(description = "排序方式：relevance/time/hot") @RequestParam(value = "sort", defaultValue = "relevance") String sort,
            @Parameter(description = "页码（从1开始）") @RequestParam(value = "page", defaultValue = "1") int page,
            @Parameter(description = "每页大小") @RequestParam(value = "size", defaultValue = "20") int size) {

        try {
            if (size > 100) {
                size = 100;
            }

            Page<VideoSearchVO> result = videoSearchService.search(keyword, categoryId, tag, userId, sort, page, size);

            Map<String, Object> data = new HashMap<>();
            data.put("content", result.getContent());
            data.put("totalElements", result.getTotalElements());
            data.put("totalPages", result.getTotalPages());
            data.put("page", result.getNumber());
            data.put("size", result.getSize());
            data.put("hasNext", result.hasNext());
            data.put("hasPrevious", result.hasPrevious());

            return Result.success("搜索成功", data);
        } catch (Exception e) {
            return Result.error("搜索失败: " + e.getMessage());
        }
    }

    @GetMapping("/suggest")
    @Operation(summary = "搜索建议", description = "根据输入的关键词获取搜索建议")
    public Result<List<String>> suggest(
            @Parameter(description = "搜索关键词") @RequestParam("keyword") String keyword) {

        try {
            List<String> suggestions = videoSearchService.suggest(keyword);
            return Result.success("获取成功", suggestions);
        } catch (Exception e) {
            return Result.error("获取搜索建议失败: " + e.getMessage());
        }
    }
}