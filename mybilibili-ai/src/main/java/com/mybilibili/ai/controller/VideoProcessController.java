package com.mybilibili.ai.controller;

import com.mybilibili.ai.audit.AiAuditLogRecorder;
import com.mybilibili.ai.mapper.VideoMapper;
import com.mybilibili.ai.process.ProcessMode;
import com.mybilibili.ai.process.StepExecutionResult;
import com.mybilibili.ai.process.VideoProcessContext;
import com.mybilibili.ai.process.VideoProcessOrchestrator;
import com.mybilibili.ai.process.VideoProcessStepType;
import com.mybilibili.common.entity.Video;
import com.mybilibili.common.operation.OperationTaskRecorder;
import com.mybilibili.common.vo.Result;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/ai/process")
@Tag(name = "视频处理接口", description = "手动触发视频处理")
public class VideoProcessController {

    @Autowired
    private VideoMapper videoMapper;

    @Autowired
    private VideoProcessOrchestrator orchestrator;

    @Autowired
    private AiAuditLogRecorder auditLogRecorder;

    @Autowired
    private OperationTaskRecorder operationTaskRecorder;

    @PostMapping("/transcode/{videoId}")
    @Operation(summary = "手动转码", description = "手动触发视频转码")
    public Result<Void> transcode(@PathVariable Integer videoId,
                                   @RequestParam Integer manuscriptId,
                                   HttpServletRequest request) {
        operationTaskRecorder.running(request, videoTaskKey("transcode", videoId), "VIDEO_PROCESS", "手动视频转码",
                "video", String.valueOf(videoId), 0, "TRANSCODING", "视频转码执行中");
        Result<Void> result = executeWithoutData(videoId, manuscriptId, null, null, VideoProcessStepType.TRANSCODE);
        operationTaskRecorder.recordResult(request, videoTaskKey("transcode", videoId), "VIDEO_PROCESS", "手动视频转码",
                "video", String.valueOf(videoId), "TRANSCODING", result);
        auditLogRecorder.record(request, "video_transcode_execute", "video", String.valueOf(videoId), result);
        return result;
    }

    @PostMapping("/audio/{videoId}")
    @Operation(summary = "手动提取音频", description = "手动触发音频提取")
    public Result<Void> extractAudio(@PathVariable Integer videoId,
                                     @RequestParam Integer manuscriptId,
                                     HttpServletRequest request) {
        operationTaskRecorder.running(request, videoTaskKey("audio", videoId), "VIDEO_PROCESS", "手动音频提取",
                "video", String.valueOf(videoId), 0, "AUDIO_EXTRACTING", "音频提取执行中");
        Result<Void> result = executeWithoutData(videoId, manuscriptId, null, null, VideoProcessStepType.EXTRACT_AUDIO);
        operationTaskRecorder.recordResult(request, videoTaskKey("audio", videoId), "VIDEO_PROCESS", "手动音频提取",
                "video", String.valueOf(videoId), "AUDIO_EXTRACTING", result);
        auditLogRecorder.record(request, "video_audio_execute", "video", String.valueOf(videoId), result);
        return result;
    }

    @PostMapping("/subtitle/{videoId}")
    @Operation(summary = "手动生成字幕", description = "手动触发字幕生成")
    public Result<Void> generateSubtitle(@PathVariable Integer videoId,
                                          @RequestParam Integer manuscriptId,
                                          HttpServletRequest request) {
        operationTaskRecorder.running(request, videoTaskKey("subtitle", videoId), "VIDEO_PROCESS", "手动字幕生成",
                "video", String.valueOf(videoId), 0, "SUBTITLE_GENERATING", "字幕生成执行中");
        Result<Void> result = executeWithoutData(videoId, manuscriptId, null, null, VideoProcessStepType.GENERATE_SUBTITLE);
        operationTaskRecorder.recordResult(request, videoTaskKey("subtitle", videoId), "VIDEO_PROCESS", "手动字幕生成",
                "video", String.valueOf(videoId), "SUBTITLE_GENERATING", result);
        auditLogRecorder.record(request, "video_subtitle_execute", "video", String.valueOf(videoId), result);
        return result;
    }

    @PostMapping("/ai-summary/{videoId}")
    @Operation(summary = "手动AI总结", description = "手动触发AI总结")
    public Result<String> aiSummary(@PathVariable Integer videoId,
                                   @RequestParam Integer manuscriptId,
                                   @RequestParam(required = false) String videoTitle,
                                   @RequestParam(required = false) String videoDescription,
                                   HttpServletRequest request) {
        operationTaskRecorder.running(request, videoTaskKey("ai-summary", videoId), "VIDEO_PROCESS", "手动AI总结",
                "video", String.valueOf(videoId), 0, "AI_SUMMARIZING", "AI总结执行中");
        Result<String> result = executeWithData(videoId, manuscriptId, videoTitle, videoDescription, VideoProcessStepType.AI_SUMMARY);
        operationTaskRecorder.recordResult(request, videoTaskKey("ai-summary", videoId), "VIDEO_PROCESS", "手动AI总结",
                "video", String.valueOf(videoId), "AI_SUMMARIZING", result);
        auditLogRecorder.record(request, "video_ai_summary_execute", "video", String.valueOf(videoId), result);
        return result;
    }

    private Result<Void> executeWithoutData(Integer videoId,
                                            Integer manuscriptId,
                                            String videoTitle,
                                            String videoDescription,
                                            VideoProcessStepType stepType) {
        try {
            StepExecutionResult result = executeStep(videoId, manuscriptId, videoTitle, videoDescription, stepType);
            if (!result.isSuccess()) {
                return Result.error(result.getErrorMessage());
            }
            return Result.success(stepType.getSuccessStageText(), null);
        } catch (Exception e) {
            return Result.error(stepType.getSuccessStageText() + "异常: " + e.getMessage());
        }
    }

    private Result<String> executeWithData(Integer videoId,
                                           Integer manuscriptId,
                                           String videoTitle,
                                           String videoDescription,
                                           VideoProcessStepType stepType) {
        try {
            StepExecutionResult result = executeStep(videoId, manuscriptId, videoTitle, videoDescription, stepType);
            if (!result.isSuccess()) {
                return Result.error(result.getErrorMessage());
            }
            Object data = result.getData();
            return Result.success(stepType.getSuccessStageText(), data == null ? null : data.toString());
        } catch (Exception e) {
            return Result.error(stepType.getSuccessStageText() + "异常: " + e.getMessage());
        }
    }

    private StepExecutionResult executeStep(Integer videoId,
                                            Integer manuscriptId,
                                            String videoTitle,
                                            String videoDescription,
                                            VideoProcessStepType stepType) {
        Video video = videoMapper.selectById(videoId);
        VideoProcessContext context = new VideoProcessContext();
        context.setVideoId(videoId);
        context.setManuscriptId(manuscriptId);
        context.setVideoTitle(videoTitle != null ? videoTitle : video != null ? video.getTitle() : "未知视频");
        context.setVideoDescription(videoDescription != null ? videoDescription : video != null ? video.getDescription() : null);
        context.setCurrentStep(stepType);
        context.setProcessMode(ProcessMode.MANUAL_SINGLE);
        context.setTriggerSource("MANUAL_API");
        return orchestrator.executeStep(context);
    }

    private String videoTaskKey(String action, Integer videoId) {
        return "video-process:" + action + ":" + videoId;
    }
}
