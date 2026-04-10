package com.mybilibili.video.controller;

import com.mybilibili.common.vo.ManuscriptVO;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.entity.Video;
import com.mybilibili.video.feign.UserClient;
import com.mybilibili.video.feign.VideoProcessClient;
import com.mybilibili.video.service.ManuscriptService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/manuscript/admin")
@Tag(name = "稿件管理接口", description = "管理员稿件审核、发布、管理等功能")
public class ManuscriptAdminController {

    @Autowired
    private ManuscriptService manuscriptService;

    @Autowired
    private UserClient userClient;

    @Autowired
    private VideoProcessClient videoProcessClient;

    @GetMapping("/pending")
    @Operation(summary = "获取待审核稿件", description = "获取所有待审核的稿件列表")
    public Result<List<ManuscriptVO>> getPendingManuscripts() {
        List<ManuscriptVO> manuscripts = manuscriptService.getPendingManuscripts();
        return Result.success("获取成功", manuscripts);
    }

    @GetMapping("/processing")
    @Operation(summary = "获取处理中稿件", description = "获取所有处理中的稿件列表")
    public Result<List<ManuscriptVO>> getProcessingManuscripts() {
        List<ManuscriptVO> manuscripts = manuscriptService.getProcessingManuscripts();
        return Result.success("获取成功", manuscripts);
    }

    @GetMapping("/ready")
    @Operation(summary = "获取待上架稿件", description = "获取所有待上架的稿件列表")
    public Result<List<ManuscriptVO>> getReadyManuscripts() {
        List<ManuscriptVO> manuscripts = manuscriptService.getReadyManuscripts();
        return Result.success("获取成功", manuscripts);
    }

    @GetMapping("/all")
    @Operation(summary = "获取所有稿件", description = "获取所有稿件列表")
    public Result<List<ManuscriptVO>> getAllManuscripts() {
        List<ManuscriptVO> manuscripts = manuscriptService.getAllManuscripts();
        return Result.success("获取成功", manuscripts);
    }

    @GetMapping("/statistics")
    @Operation(summary = "获取稿件统计", description = "获取各状态稿件数量统计")
    public Result<Map<String, Object>> getManuscriptStatistics() {
        Map<String, Object> statistics = manuscriptService.getManuscriptStatistics();
        return Result.success("获取成功", statistics);
    }

    @GetMapping("/{id}")
    @Operation(summary = "获取稿件详情", description = "根据ID获取稿件详情，包含视频列表")
    public Result<Map<String, Object>> getManuscriptDetail(@PathVariable Integer id) {
        Map<String, Object> detail = manuscriptService.getManuscriptDetail(id);
        if (detail == null) {
            return Result.error(404, "稿件不存在");
        }
        return Result.success("获取成功", detail);
    }

    @GetMapping("/{id}/videos")
    @Operation(summary = "获取稿件视频列表", description = "根据稿件ID获取该稿件的所有视频")
    public Result<List<ManuscriptVO.VideoItemVO>> getManuscriptVideos(@PathVariable Integer id) {
        List<ManuscriptVO.VideoItemVO> videos = manuscriptService.getManuscriptVideos(id);
        return Result.success("获取成功", videos);
    }

    @PostMapping("/approve/{id}")
    @Operation(summary = "审核通过", description = "审核通过稿件，状态改为处理中")
    public Result<Void> approveManuscript(
            @PathVariable Integer id,
            @Parameter(description = "审核员ID") @RequestParam(required = false) Integer reviewerId,
            @Parameter(description = "审核理由") @RequestParam(required = false) String reason) {
        boolean success = manuscriptService.approveManuscript(id, reviewerId, reason);
        if (success) {
            return Result.success("审核通过", null);
        } else {
            return Result.error("稿件不存在或审核失败");
        }
    }

    @PostMapping("/reject/{id}")
    @Operation(summary = "审核拒绝", description = "审核拒绝稿件")
    public Result<Void> rejectManuscript(
            @PathVariable Integer id,
            @Parameter(description = "审核员ID") @RequestParam(required = false) Integer reviewerId,
            @Parameter(description = "拒绝理由") @RequestParam(required = false) String reason) {
        boolean success = manuscriptService.rejectManuscript(id, reviewerId, reason);
        if (success) {
            return Result.success("审核拒绝成功", null);
        } else {
            return Result.error("稿件不存在或审核失败");
        }
    }

    @PostMapping("/publish/{id}")
    @Operation(summary = "发布稿件", description = "发布稿件到前台")
    public Result<Void> publishManuscript(@PathVariable Integer id) {
        boolean success = manuscriptService.publishManuscript(id);
        if (success) {
            return Result.success("发布成功", null);
        } else {
            return Result.error("稿件不存在或发布失败");
        }
    }

    @PostMapping("/unpublish/{id}")
    @Operation(summary = "下架稿件", description = "下架前台稿件")
    public Result<Void> unpublishManuscript(@PathVariable Integer id) {
        boolean success = manuscriptService.unpublishManuscript(id);
        if (success) {
            return Result.success("下架成功", null);
        } else {
            return Result.error("稿件不存在或下架失败");
        }
    }

    @PostMapping("/retry/{id}")
    @Operation(summary = "重试发布", description = "重新尝试发布失败的稿件")
    public Result<Void> retryPublish(@PathVariable Integer id) {
        boolean success = manuscriptService.retryManuscript(id);
        if (success) {
            return Result.success("重试发布成功", null);
        } else {
            return Result.error("稿件不存在或重试失败");
        }
    }

    @PostMapping("/transcode/{videoId}")
    @Operation(summary = "手动开始转码", description = "手动触发视频转码")
    public Result<Void> manualTranscode(@PathVariable Integer videoId) {
        Video video = manuscriptService.getVideoById(videoId);
        if (video == null) {
            return Result.error("视频不存在");
        }
        try {
            return videoProcessClient.triggerTranscode(videoId, video.getManuscriptId());
        } catch (Exception e) {
            return Result.error("调用转码服务失败: " + e.getMessage());
        }
    }

    @PostMapping("/extract-audio/{videoId}")
    @Operation(summary = "手动提取音频", description = "手动触发音频提取")
    public Result<Void> manualExtractAudio(@PathVariable Integer videoId) {
        Video video = manuscriptService.getVideoById(videoId);
        if (video == null) {
            return Result.error("视频不存在");
        }
        try {
            return videoProcessClient.triggerAudioExtract(videoId, video.getManuscriptId());
        } catch (Exception e) {
            return Result.error("调用音频提取服务失败: " + e.getMessage());
        }
    }

    @PostMapping("/generate-subtitle/{videoId}")
    @Operation(summary = "手动生成字幕", description = "手动触发字幕生成")
    public Result<Void> manualGenerateSubtitle(@PathVariable Integer videoId) {
        Video video = manuscriptService.getVideoById(videoId);
        if (video == null) {
            return Result.error("视频不存在");
        }
        try {
            return videoProcessClient.triggerSubtitleGenerate(videoId, video.getManuscriptId());
        } catch (Exception e) {
            return Result.error("调用字幕生成服务失败: " + e.getMessage());
        }
    }

    @PostMapping("/ai-summary/{videoId}")
    @Operation(summary = "手动AI总结", description = "手动触发AI总结")
    public Result<Void> manualAiSummary(@PathVariable Integer videoId) {
        Video video = manuscriptService.getVideoById(videoId);
        if (video == null) {
            return Result.error("视频不存在");
        }
        try {
            return videoProcessClient.triggerAiSummary(videoId, video.getManuscriptId());
        } catch (Exception e) {
            return Result.error("调用AI总结服务失败: " + e.getMessage());
        }
    }

    @PostMapping("/process-all/{videoId}")
    @Operation(summary = "一键处理视频", description = "一键执行所有处理步骤")
    public Result<Void> manualProcessAll(@PathVariable Integer videoId) {
        boolean success = manuscriptService.manualProcessAll(videoId);
        if (success) {
            return Result.success("已开始全流程处理", null);
        } else {
            return Result.error("视频不存在或处理失败");
        }
    }

    @PostMapping("/reset/{videoId}")
    @Operation(summary = "重置视频状态", description = "重置视频处理状态")
    public Result<Void> resetVideoStatus(@PathVariable Integer videoId) {
        boolean success = manuscriptService.resetVideoStatus(videoId);
        if (success) {
            return Result.success("重置成功", null);
        } else {
            return Result.error("视频不存在或重置失败");
        }
    }
}
