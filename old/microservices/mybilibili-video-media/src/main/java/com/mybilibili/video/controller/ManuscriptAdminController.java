package com.mybilibili.video.controller;

import com.mybilibili.common.audit.AuditLogRecorder;
import com.mybilibili.common.vo.ManuscriptVO;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.entity.Video;
import com.mybilibili.common.operation.OperationTaskRecorder;
import com.mybilibili.video.feign.UserClient;
import com.mybilibili.video.service.ManuscriptService;
import com.mybilibili.video.service.VideoProcessPort;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import jakarta.servlet.http.HttpServletRequest;
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
    private VideoProcessPort videoProcessPort;

    @Autowired
    private AuditLogRecorder auditLogRecorder;

    @Autowired
    private OperationTaskRecorder operationTaskRecorder;

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
            @Parameter(description = "审核理由") @RequestParam(required = false) String reason,
            HttpServletRequest request) {
        boolean success = manuscriptService.approveManuscript(id, reviewerId, reason);
        Result<Void> result;
        if (success) {
            result = Result.success("审核通过", null);
        } else {
            result = Result.error("稿件不存在或审核失败");
        }
        auditLogRecorder.record(request, "manuscript", "manuscript_approve", "manuscript", String.valueOf(id), result);
        return result;
    }

    @PostMapping("/{id}/approve-with-process")
    @Operation(summary = "审核通过（可选自动处理）", description = "审核通过稿件，可选择是否自动处理视频")
    public Result<Map<String, Object>> approveManuscriptWithProcess(
            @PathVariable Integer id,
            @Parameter(description = "是否自动处理视频") @RequestParam(defaultValue = "false") boolean autoProcess,
            @Parameter(description = "审核员ID") @RequestParam(required = false) Integer reviewerId,
            @Parameter(description = "审核理由") @RequestParam(required = false) String reason,
            HttpServletRequest request) {
        
        boolean success = manuscriptService.approveManuscriptWithProcess(id, reviewerId, reason, autoProcess);
        if (success) {
            Map<String, Object> data = new java.util.HashMap<>();
            data.put("manuscriptId", id);
            data.put("autoProcess", autoProcess);
            data.put("message", autoProcess ? "审核通过，已提交全流程处理任务" : "审核通过");
            Result<Map<String, Object>> response = Result.success("审核成功", data);
            auditLogRecorder.record(request, "manuscript", autoProcess ? "manuscript_approve_with_process" : "manuscript_approve",
                    "manuscript", String.valueOf(id), response);
            return response;
        } else {
            Result<Map<String, Object>> result = Result.error("稿件不存在或审核失败");
            auditLogRecorder.record(request, "manuscript", autoProcess ? "manuscript_approve_with_process" : "manuscript_approve",
                    "manuscript", String.valueOf(id), result);
            return result;
        }
    }

    @PostMapping("/reject/{id}")
    @Operation(summary = "审核拒绝", description = "审核拒绝稿件")
    public Result<Void> rejectManuscript(
            @PathVariable Integer id,
            @Parameter(description = "审核员ID") @RequestParam(required = false) Integer reviewerId,
            @Parameter(description = "拒绝理由") @RequestParam(required = false) String reason,
            HttpServletRequest request) {
        boolean success = manuscriptService.rejectManuscript(id, reviewerId, reason);
        Result<Void> result;
        if (success) {
            result = Result.success("审核拒绝成功", null);
        } else {
            result = Result.error("稿件不存在或审核失败");
        }
        auditLogRecorder.record(request, "manuscript", "manuscript_reject", "manuscript", String.valueOf(id), result);
        return result;
    }

    @PostMapping("/publish/{id}")
    @Operation(summary = "发布稿件", description = "发布稿件到前台")
    public Result<Void> publishManuscript(@PathVariable Integer id, HttpServletRequest request) {
        boolean success = manuscriptService.publishManuscript(id);
        Result<Void> result;
        if (success) {
            result = Result.success("发布成功", null);
        } else {
            result = Result.error("稿件不存在或发布失败");
        }
        auditLogRecorder.record(request, "manuscript", "manuscript_publish", "manuscript", String.valueOf(id), result);
        return result;
    }

    @PostMapping("/unpublish/{id}")
    @Operation(summary = "下架稿件", description = "下架前台稿件")
    public Result<Void> unpublishManuscript(@PathVariable Integer id, HttpServletRequest request) {
        boolean success = manuscriptService.unpublishManuscript(id);
        Result<Void> result;
        if (success) {
            result = Result.success("下架成功", null);
        } else {
            result = Result.error("稿件不存在或下架失败");
        }
        auditLogRecorder.record(request, "manuscript", "manuscript_unpublish", "manuscript", String.valueOf(id), result);
        return result;
    }

    @PostMapping("/retry/{id}")
    @Operation(summary = "重试发布", description = "重新尝试发布失败的稿件")
    public Result<Void> retryPublish(@PathVariable Integer id, HttpServletRequest request) {
        boolean success = manuscriptService.retryManuscript(id);
        Result<Void> result;
        if (success) {
            result = Result.success("重试发布成功", null);
        } else {
            result = Result.error("稿件不存在或重试失败");
        }
        auditLogRecorder.record(request, "manuscript", "manuscript_retry_publish", "manuscript", String.valueOf(id), result);
        return result;
    }

    @PostMapping("/transcode/{videoId}")
    @Operation(summary = "手动开始转码", description = "手动触发视频转码")
    public Result<Void> manualTranscode(@PathVariable Integer videoId, HttpServletRequest request) {
        Video video = manuscriptService.getVideoById(videoId);
        if (video == null) {
            Result<Void> result = Result.error("视频不存在");
            operationTaskRecorder.failed(request, videoTaskKey("transcode", videoId), "VIDEO_PROCESS", "手动视频转码",
                    "video", String.valueOf(videoId), 0, "TRANSCODING", result.getMessage(), result.getMessage());
            auditLogRecorder.record(request, "task", "video_transcode_trigger", "video", String.valueOf(videoId), result);
            return result;
        }
        try {
            operationTaskRecorder.running(request, videoTaskKey("transcode", videoId), "VIDEO_PROCESS", "手动视频转码",
                    "video", String.valueOf(videoId), 0, "TRANSCODING", "转码任务已触发");
            Result<Void> result = videoProcessPort.triggerTranscode(videoId, video.getManuscriptId());
            operationTaskRecorder.recordResult(request, videoTaskKey("transcode", videoId), "VIDEO_PROCESS", "手动视频转码",
                    "video", String.valueOf(videoId), "TRANSCODING", result);
            auditLogRecorder.record(request, "task", "video_transcode_trigger", "video", String.valueOf(videoId), result);
            return result;
        } catch (Exception e) {
            Result<Void> result = Result.error("调用视频媒体处理失败: " + e.getMessage());
            operationTaskRecorder.failed(request, videoTaskKey("transcode", videoId), "VIDEO_PROCESS", "手动视频转码",
                    "video", String.valueOf(videoId), 0, "TRANSCODING", result.getMessage(), e.getMessage());
            auditLogRecorder.record(request, "task", "video_transcode_trigger", "video", String.valueOf(videoId), result);
            return result;
        }
    }

    @PostMapping("/extract-audio/{videoId}")
    @Operation(summary = "手动提取音频", description = "手动触发音频提取")
    public Result<Void> manualExtractAudio(@PathVariable Integer videoId, HttpServletRequest request) {
        Video video = manuscriptService.getVideoById(videoId);
        if (video == null) {
            Result<Void> result = Result.error("视频不存在");
            operationTaskRecorder.failed(request, videoTaskKey("audio", videoId), "VIDEO_PROCESS", "手动音频提取",
                    "video", String.valueOf(videoId), 0, "AUDIO_EXTRACTING", result.getMessage(), result.getMessage());
            auditLogRecorder.record(request, "task", "video_audio_trigger", "video", String.valueOf(videoId), result);
            return result;
        }
        try {
            operationTaskRecorder.running(request, videoTaskKey("audio", videoId), "VIDEO_PROCESS", "手动音频提取",
                    "video", String.valueOf(videoId), 0, "AUDIO_EXTRACTING", "音频提取任务已触发");
            Result<Void> result = videoProcessPort.triggerAudioExtract(videoId, video.getManuscriptId());
            operationTaskRecorder.recordResult(request, videoTaskKey("audio", videoId), "VIDEO_PROCESS", "手动音频提取",
                    "video", String.valueOf(videoId), "AUDIO_EXTRACTING", result);
            auditLogRecorder.record(request, "task", "video_audio_trigger", "video", String.valueOf(videoId), result);
            return result;
        } catch (Exception e) {
            Result<Void> result = Result.error("调用视频媒体处理失败: " + e.getMessage());
            operationTaskRecorder.failed(request, videoTaskKey("audio", videoId), "VIDEO_PROCESS", "手动音频提取",
                    "video", String.valueOf(videoId), 0, "AUDIO_EXTRACTING", result.getMessage(), e.getMessage());
            auditLogRecorder.record(request, "task", "video_audio_trigger", "video", String.valueOf(videoId), result);
            return result;
        }
    }

    @PostMapping("/generate-subtitle/{videoId}")
    @Operation(summary = "手动生成字幕", description = "手动触发字幕生成")
    public Result<Void> manualGenerateSubtitle(@PathVariable Integer videoId, HttpServletRequest request) {
        Video video = manuscriptService.getVideoById(videoId);
        if (video == null) {
            Result<Void> result = Result.error("视频不存在");
            operationTaskRecorder.failed(request, videoTaskKey("subtitle", videoId), "VIDEO_PROCESS", "手动字幕生成",
                    "video", String.valueOf(videoId), 0, "SUBTITLE_GENERATING", result.getMessage(), result.getMessage());
            auditLogRecorder.record(request, "task", "video_subtitle_trigger", "video", String.valueOf(videoId), result);
            return result;
        }
        try {
            operationTaskRecorder.running(request, videoTaskKey("subtitle", videoId), "VIDEO_PROCESS", "手动字幕生成",
                    "video", String.valueOf(videoId), 0, "SUBTITLE_GENERATING", "字幕生成任务已触发");
            Result<Void> result = videoProcessPort.triggerSubtitleGenerate(videoId, video.getManuscriptId());
            operationTaskRecorder.recordResult(request, videoTaskKey("subtitle", videoId), "VIDEO_PROCESS", "手动字幕生成",
                    "video", String.valueOf(videoId), "SUBTITLE_GENERATING", result);
            auditLogRecorder.record(request, "task", "video_subtitle_trigger", "video", String.valueOf(videoId), result);
            return result;
        } catch (Exception e) {
            Result<Void> result = Result.error("调用字幕生成服务失败: " + e.getMessage());
            operationTaskRecorder.failed(request, videoTaskKey("subtitle", videoId), "VIDEO_PROCESS", "手动字幕生成",
                    "video", String.valueOf(videoId), 0, "SUBTITLE_GENERATING", result.getMessage(), e.getMessage());
            auditLogRecorder.record(request, "task", "video_subtitle_trigger", "video", String.valueOf(videoId), result);
            return result;
        }
    }

    @PostMapping("/ai-summary/{videoId}")
    @Operation(summary = "手动AI总结", description = "手动触发AI总结")
    public Result<Void> manualAiSummary(@PathVariable Integer videoId, HttpServletRequest request) {
        Video video = manuscriptService.getVideoById(videoId);
        if (video == null) {
            Result<Void> result = Result.error("视频不存在");
            operationTaskRecorder.failed(request, videoTaskKey("ai-summary", videoId), "VIDEO_PROCESS", "手动AI总结",
                    "video", String.valueOf(videoId), 0, "AI_SUMMARIZING", result.getMessage(), result.getMessage());
            auditLogRecorder.record(request, "task", "video_ai_summary_trigger", "video", String.valueOf(videoId), result);
            return result;
        }
        try {
            operationTaskRecorder.running(request, videoTaskKey("ai-summary", videoId), "VIDEO_PROCESS", "手动AI总结",
                    "video", String.valueOf(videoId), 0, "AI_SUMMARIZING", "AI总结任务已触发");
            Result<Void> result = videoProcessPort.triggerAiSummary(videoId, video.getManuscriptId());
            operationTaskRecorder.recordResult(request, videoTaskKey("ai-summary", videoId), "VIDEO_PROCESS", "手动AI总结",
                    "video", String.valueOf(videoId), "AI_SUMMARIZING", result);
            auditLogRecorder.record(request, "task", "video_ai_summary_trigger", "video", String.valueOf(videoId), result);
            return result;
        } catch (Exception e) {
            Result<Void> result = Result.error("调用AI总结服务失败: " + e.getMessage());
            operationTaskRecorder.failed(request, videoTaskKey("ai-summary", videoId), "VIDEO_PROCESS", "手动AI总结",
                    "video", String.valueOf(videoId), 0, "AI_SUMMARIZING", result.getMessage(), e.getMessage());
            auditLogRecorder.record(request, "task", "video_ai_summary_trigger", "video", String.valueOf(videoId), result);
            return result;
        }
    }

    @PostMapping("/process-all/{videoId}")
    @Operation(summary = "一键处理视频", description = "一键执行所有处理步骤")
    public Result<Void> manualProcessAll(@PathVariable Integer videoId, HttpServletRequest request) {
        operationTaskRecorder.running(request, videoTaskKey("process-all", videoId), "VIDEO_PROCESS", "一键视频处理",
                "video", String.valueOf(videoId), 0, "PROCESS_ALL", "全流程处理任务已触发");
        boolean success = manuscriptService.manualProcessAll(videoId);
        Result<Void> result;
        if (success) {
            result = Result.success("已开始全流程处理", null);
            operationTaskRecorder.success(request, videoTaskKey("process-all", videoId), "VIDEO_PROCESS", "一键视频处理",
                    "video", String.valueOf(videoId), "PROCESS_ALL", result.getMessage());
        } else {
            result = Result.error("视频不存在或处理失败");
            operationTaskRecorder.failed(request, videoTaskKey("process-all", videoId), "VIDEO_PROCESS", "一键视频处理",
                    "video", String.valueOf(videoId), 0, "PROCESS_ALL", result.getMessage(), result.getMessage());
        }
        auditLogRecorder.record(request, "task", "video_process_all_trigger", "video", String.valueOf(videoId), result);
        return result;
    }

    @GetMapping("/video-source/{videoId}")
    @Operation(summary = "获取视频播放地址", description = "根据视频ID获取视频源地址用于播放")
    public Result<Map<String, Object>> getVideoSourceUrl(@PathVariable Integer videoId) {
        Map<String, Object> result = manuscriptService.getVideoSourceUrl(videoId);
        if (result == null) {
            return Result.error(404, "视频不存在");
        }
        return Result.success("获取成功", result);
    }

    @PostMapping("/reset/{videoId}")
    @Operation(summary = "重置视频状态", description = "重置视频处理状态")
    public Result<Void> resetVideoStatus(@PathVariable Integer videoId, HttpServletRequest request) {
        boolean success = manuscriptService.resetVideoStatus(videoId);
        Result<Void> result;
        if (success) {
            result = Result.success("重置成功", null);
            operationTaskRecorder.success(request, videoTaskKey("reset", videoId), "VIDEO_PROCESS", "重置视频处理",
                    "video", String.valueOf(videoId), "RESET", result.getMessage());
        } else {
            result = Result.error("视频不存在或重置失败");
            operationTaskRecorder.failed(request, videoTaskKey("reset", videoId), "VIDEO_PROCESS", "重置视频处理",
                    "video", String.valueOf(videoId), 0, "RESET", result.getMessage(), result.getMessage());
        }
        auditLogRecorder.record(request, "task", "video_process_reset", "video", String.valueOf(videoId), result);
        return result;
    }

    private String videoTaskKey(String action, Integer videoId) {
        return "video-process:" + action + ":" + videoId;
    }
}
