package com.mybilibili.ai.controller;

import com.mybilibili.ai.config.UploadFilePathConfig;
import com.mybilibili.ai.mapper.VideoMapper;
import com.mybilibili.ai.repository.SubtitleRepository;
import com.mybilibili.ai.service.AiSubtitleService;
import com.mybilibili.ai.service.AiSummaryService;
import com.mybilibili.ai.service.AudioExtractService;
import com.mybilibili.ai.service.VideoTranscodeService;
import com.mybilibili.common.entity.Subtitle;
import com.mybilibili.common.entity.Video;
import com.mybilibili.common.vo.Result;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/ai/process")
@Tag(name = "视频处理接口", description = "手动触发视频处理")
public class VideoProcessController {

    @Autowired
    private VideoTranscodeService videoTranscodeService;

    @Autowired
    private AudioExtractService audioExtractService;

    @Autowired
    private AiSubtitleService aiSubtitleService;

    @Autowired
    private AiSummaryService aiSummaryService;

    @Autowired
    private SubtitleRepository subtitleRepository;

    @Autowired
    private VideoMapper videoMapper;

    @Autowired
    private UploadFilePathConfig uploadFilePathConfig;

    @Autowired
    private StringRedisTemplate redisTemplate;

    @PostMapping("/transcode/{videoId}")
    @Operation(summary = "手动转码", description = "手动触发视频转码")
    public Result<Void> transcode(@PathVariable Integer videoId,
                                   @RequestParam Integer manuscriptId) {
        try {
            // 先设置状态为转码中(1)
            videoMapper.updateProcessStatus(videoId, Video.PROCESS_STATUS_TRANSCODING);

            boolean success = videoTranscodeService.transcode(manuscriptId, videoId);
            if (success) {
                // 转码成功，设置状态为转码成功(11)
                videoMapper.updateTranscodeResult(
                    videoId,
                    Video.PROCESS_STATUS_TRANSCODE_SUCCESS,
                    "/uploads/manuscripts/" + manuscriptId + "/videos/" + videoId + "/transcoded/1080p.mp4",
                    "/uploads/manuscripts/" + manuscriptId + "/videos/" + videoId + "/transcoded/720p.mp4",
                    "/uploads/manuscripts/" + manuscriptId + "/videos/" + videoId + "/transcoded/480p.mp4"
                );
                return Result.success("转码成功", null);
            } else {
                // 转码失败，设置状态为转码失败(10)
                videoMapper.updateProcessStatus(videoId, Video.PROCESS_STATUS_TRANSCODE_FAILED);
                return Result.error("转码失败");
            }
        } catch (Exception e) {
            // 转码异常，设置状态为转码失败(10)
            videoMapper.updateProcessStatus(videoId, Video.PROCESS_STATUS_TRANSCODE_FAILED);
            return Result.error("转码异常: " + e.getMessage());
        }
    }

    @PostMapping("/audio/{videoId}")
    @Operation(summary = "手动提取音频", description = "手动触发音频提取")
    public Result<Void> extractAudio(@PathVariable Integer videoId,
                                     @RequestParam Integer manuscriptId) {
        try {
            // 先设置状态为音频提取中(2)
            videoMapper.updateProcessStatus(videoId, Video.PROCESS_STATUS_AUDIO_EXTRACTING);

            boolean success = audioExtractService.extractAudio(manuscriptId, videoId);
            if (success) {
                videoMapper.updateProcessStatus(videoId, Video.PROCESS_STATUS_AUDIO_SUCCESS);
                return Result.success("音频提取成功", null);
            } else {
                videoMapper.updateProcessStatus(videoId, Video.PROCESS_STATUS_AUDIO_FAILED);
                return Result.error("音频提取失败");
            }
        } catch (Exception e) {
            videoMapper.updateProcessStatus(videoId, Video.PROCESS_STATUS_AUDIO_FAILED);
            return Result.error("音频提取异常: " + e.getMessage());
        }
    }

    @PostMapping("/subtitle/{videoId}")
    @Operation(summary = "手动生成字幕", description = "手动触发字幕生成")
    public Result<Void> generateSubtitle(@PathVariable Integer videoId,
                                          @RequestParam Integer manuscriptId) {
        try {
            // 先设置状态为字幕生成中(3)
            videoMapper.updateProcessStatus(videoId, Video.PROCESS_STATUS_SUBTITLE_GENERATING);

            boolean success = aiSubtitleService.generateSubtitle(manuscriptId, videoId);
            if (success) {
                videoMapper.updateProcessStatus(videoId, Video.PROCESS_STATUS_SUBTITLE_SUCCESS);
                return Result.success("字幕生成成功", null);
            } else {
                videoMapper.updateProcessStatus(videoId, Video.PROCESS_STATUS_SUBTITLE_FAILED);
                return Result.error("字幕生成失败");
            }
        } catch (Exception e) {
            videoMapper.updateProcessStatus(videoId, Video.PROCESS_STATUS_SUBTITLE_FAILED);
            return Result.error("字幕生成异常: " + e.getMessage());
        }
    }

    @PostMapping("/ai-summary/{videoId}")
    @Operation(summary = "手动AI总结", description = "手动触发AI总结")
    public Result<String> aiSummary(@PathVariable Integer videoId,
                                   @RequestParam Integer manuscriptId,
                                   @RequestParam(required = false) String videoTitle,
                                   @RequestParam(required = false) String videoDescription) {
        try {
            videoMapper.updateProcessStatus(videoId, Video.PROCESS_STATUS_AI_SUMMARIZING);

            List<Subtitle> subtitles = subtitleRepository.findByVideoId(videoId);
            if (subtitles == null || subtitles.isEmpty()) {
                videoMapper.updateProcessStatus(videoId, Video.PROCESS_STATUS_AI_FAILED);
                return Result.error("请先生成字幕");
            }

            StringBuilder subtitleText = new StringBuilder();
            for (Subtitle subtitle : subtitles) {
                if (subtitle.getContent() != null) {
                    for (Subtitle.SubtitleItem item : subtitle.getContent()) {
                        if (item.getText() != null) {
                            subtitleText.append(item.getText()).append("\n");
                        }
                    }
                }
            }

            String summary = aiSummaryService.generateSummary(
                subtitleText.toString(),
                videoTitle != null ? videoTitle : "视频",
                videoDescription != null ? videoDescription : ""
            );

            if (summary != null && !summary.isEmpty()) {
                String summaryPath = uploadFilePathConfig.getAiSummaryPath(manuscriptId, videoId);
                aiSummaryService.saveSummaryToFile(summary, summaryPath, videoTitle != null ? videoTitle : "视频");

                videoMapper.updateProcessStatus(videoId, Video.PROCESS_STATUS_COMPLETED);
                videoMapper.updateHasSummary(videoId, 1);

                return Result.success("AI总结成功", summary);
            } else {
                videoMapper.updateProcessStatus(videoId, Video.PROCESS_STATUS_AI_FAILED);
                return Result.error("AI总结失败");
            }
        } catch (Exception e) {
            videoMapper.updateProcessStatus(videoId, Video.PROCESS_STATUS_AI_FAILED);
            return Result.error("AI总结异常: " + e.getMessage());
        }
    }
}
