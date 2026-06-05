package com.mybilibili.video.controller;

import com.mybilibili.common.vo.Result;
import com.mybilibili.video.service.VideoService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/video/admin")
@Tag(name = "视频管理接口", description = "视频列表查询、详情查看、删除等管理功能")
public class VideoAdminController {

    @Autowired
    private VideoService videoService;

    @GetMapping("/list")
    @Operation(summary = "获取视频列表", description = "获取视频列表，支持分页、搜索和状态筛选")
    public Result<Map<String, Object>> getVideoList(
            @Parameter(description = "页码") @RequestParam(defaultValue = "1") Integer page,
            @Parameter(description = "每页数量") @RequestParam(defaultValue = "10") Integer size,
            @Parameter(description = "搜索关键词") @RequestParam(required = false) String keyword,
            @Parameter(description = "状态") @RequestParam(required = false) Integer status) {
        return videoService.getAdminVideoList(page, size, keyword, status);
    }

    @GetMapping("/{id}")
    @Operation(summary = "获取视频详情", description = "根据ID获取视频详情")
    public Result<?> getVideoById(@PathVariable Integer id) {
        return videoService.getVideoById(id);
    }

    @DeleteMapping("/{id}")
    @Operation(summary = "删除视频", description = "删除视频")
    public Result<?> deleteVideo(@PathVariable Integer id) {
        return videoService.deleteVideo(id);
    }

    @DeleteMapping("/batch")
    @Operation(summary = "批量删除视频", description = "批量删除多个视频")
    public Result<?> deleteVideos(@RequestBody List<Integer> ids) {
        return videoService.deleteVideos(ids);
    }
}