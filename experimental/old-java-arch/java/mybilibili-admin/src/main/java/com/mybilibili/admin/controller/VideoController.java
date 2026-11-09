package com.mybilibili.admin.controller;

import com.mybilibili.admin.service.VideoService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/videos")
@Tag(name = "视频查询接口", description = "视频列表查询、详情查看、删除等功能")
public class VideoController {

    @Autowired
    private VideoService videoService;

    @GetMapping
    @Operation(summary = "获取视频列表", description = "获取视频列表，支持分页、搜索和状态筛选")
    public Object getVideoList(
            @RequestParam(defaultValue = "1") Integer page,
            @RequestParam(defaultValue = "10") Integer size,
            @RequestParam(required = false) String keyword,
            @RequestParam(required = false) Integer status) {
        return videoService.getVideoList(page, size, keyword, status);
    }

    @GetMapping("/{id}")
    @Operation(summary = "获取视频详情", description = "根据ID获取视频详情")
    public Object getVideoById(@PathVariable Integer id) {
        return videoService.getVideoById(id);
    }

    @DeleteMapping("/{id}")
    @Operation(summary = "删除视频", description = "删除视频")
    public Object deleteVideo(@PathVariable Integer id) {
        return videoService.deleteVideo(id);
    }

    @DeleteMapping("/batch")
    @Operation(summary = "批量删除视频", description = "批量删除多个视频")
    public Object deleteVideos(@RequestBody List<Integer> ids) {
        return videoService.deleteVideos(ids);
    }
}
