package com.mybilibili.video.controller;

import com.mybilibili.common.entity.Video;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.VideoVO;
import com.mybilibili.video.service.VideoService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/video")
@Tag(name = "视频管理接口", description = "视频相关操作")
public class VideoController {

    @Autowired
    private VideoService videoService;

    @GetMapping("/{id}")
    @Operation(summary = "获取视频详情", description = "根据视频ID获取视频信息")
    public Result<VideoVO> getById(@PathVariable Integer id) {
        return videoService.getById(id);
    }

    @GetMapping("/manuscript/{manuscriptId}")
    @Operation(summary = "获取稿件视频列表", description = "根据稿件ID获取视频列表")
    public Result<List<Video>> getVideosByManuscriptId(@PathVariable Integer manuscriptId) {
        return videoService.getVideosByManuscriptId(manuscriptId);
    }

    @GetMapping("/user/{userId}/ids")
    @Operation(summary = "获取用户稿件ID列表", description = "根据用户ID获取其所有稿件的ID列表")
    public Result<List<Integer>> getManuscriptIdsByUserId(@PathVariable Integer userId) {
        return videoService.getManuscriptIdsByUserId(userId);
    }

    @GetMapping("/user/{userId}/video-ids")
    @Operation(summary = "获取用户视频ID列表", description = "根据用户ID获取其所有视频的ID列表")
    public Result<List<Integer>> getVideoIdsByUserId(@PathVariable Integer userId) {
        return videoService.getVideoIdsByUserId(userId);
    }
}