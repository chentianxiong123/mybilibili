package com.mybilibili.videomedia.process;

import com.mybilibili.common.entity.Video;
import com.mybilibili.common.vo.Result;
import com.mybilibili.video.mapper.VideoMapper;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping({"/ai/admin/process", "/video/process/admin"})
public class VideoProcessAdminController {

    private final VideoMapper videoMapper;
    private final VideoProcessProgressSseService sseService;

    public VideoProcessAdminController(VideoMapper videoMapper, VideoProcessProgressSseService sseService) {
        this.videoMapper = videoMapper;
        this.sseService = sseService;
    }

    @GetMapping("/stream")
    public SseEmitter stream() {
        Map<String, Object> snapshot = new HashMap<>();
        snapshot.put("current", buildCurrentTask(videoMapper.selectProcessing()));
        snapshot.put("statistics", buildStatistics(videoMapper.selectAll()));
        snapshot.put("type", "snapshot");
        return sseService.createAdminEmitter(snapshot);
    }

    @GetMapping("/current")
    public Result<Map<String, Object>> getCurrentTask() {
        try {
            return Result.success("获取成功", buildCurrentTask(videoMapper.selectProcessing()));
        } catch (Exception e) {
            return Result.error("获取当前任务失败：" + e.getMessage());
        }
    }

    @GetMapping("/queue")
    public Result<Map<String, Object>> getQueueInfo() {
        try {
            Map<String, Object> result = new HashMap<>();
            int waitingTranscode = 0;
            int waitingAudio = 0;
            int waitingSubtitle = 0;
            int waitingAi = 0;

            List<Video> allVideos = videoMapper.selectAll();
            for (Video video : allVideos) {
                Integer status = video.getProcessStatus();
                if (status == null || status == Video.PROCESS_STATUS_PENDING) {
                    waitingTranscode++;
                } else if (status == Video.PROCESS_STATUS_TRANSCODE_SUCCESS) {
                    waitingAudio++;
                } else if (status == Video.PROCESS_STATUS_AUDIO_SUCCESS) {
                    waitingSubtitle++;
                } else if (status == Video.PROCESS_STATUS_SUBTITLE_SUCCESS) {
                    waitingAi++;
                }
            }

            int queueSize = waitingTranscode + waitingAudio + waitingSubtitle + waitingAi;
            result.put("waitingTranscode", waitingTranscode);
            result.put("waitingAudio", waitingAudio);
            result.put("waitingSubtitle", waitingSubtitle);
            result.put("waitingAi", waitingAi);
            result.put("queueSize", queueSize);
            result.put("message", "队列中有 " + queueSize + " 个待处理任务");

            return Result.success("获取成功", result);
        } catch (Exception e) {
            return Result.error("获取队列信息失败：" + e.getMessage());
        }
    }

    @GetMapping("/statistics")
    public Result<Map<String, Object>> getStatistics() {
        try {
            return Result.success("获取成功", buildStatistics(videoMapper.selectAll()));
        } catch (Exception e) {
            return Result.error("获取统计数据失败：" + e.getMessage());
        }
    }

    @GetMapping("/manuscript/{manuscriptId}/status")
    public Result<Map<String, Object>> getManuscriptProcessStatus(@PathVariable Integer manuscriptId) {
        try {
            List<Video> videos = videoMapper.selectByManuscriptId(manuscriptId);
            Map<String, Object> result = new HashMap<>();
            result.put("manuscriptId", manuscriptId);
            result.put("totalVideos", videos.size());

            if (videos.isEmpty()) {
                result.put("status", "no_videos");
                result.put("statusText", "无视频");
                result.put("processedVideos", 0);
                result.put("processingVideos", 0);
                result.put("failedVideos", 0);
                result.put("pendingVideos", 0);
                return Result.success("获取成功", result);
            }

            int processedCount = 0;
            int processingCount = 0;
            int failedCount = 0;
            int pendingCount = 0;
            for (Video video : videos) {
                Integer status = video.getProcessStatus();
                if (status == null) {
                    status = Video.PROCESS_STATUS_PENDING;
                }

                if (status == Video.PROCESS_STATUS_COMPLETED) {
                    processedCount++;
                } else if (isProcessingStatus(status)) {
                    processingCount++;
                } else if (isFailedStatus(status)) {
                    failedCount++;
                } else {
                    pendingCount++;
                }
            }

            result.put("processedVideos", processedCount);
            result.put("processingVideos", processingCount);
            result.put("failedVideos", failedCount);
            result.put("pendingVideos", pendingCount);

            String status;
            String statusText;
            if (processedCount == videos.size()) {
                status = "completed";
                statusText = "全部处理完成";
            } else if (processingCount > 0) {
                status = "processing";
                statusText = "处理中";
            } else if (failedCount > 0) {
                status = "failed";
                statusText = "处理失败";
            } else {
                status = "pending";
                statusText = "待处理";
            }

            result.put("status", status);
            result.put("statusText", statusText);
            return Result.success("获取成功", result);
        } catch (Exception e) {
            return Result.error("获取稿件处理状态失败：" + e.getMessage());
        }
    }

    private Map<String, Object> buildCurrentTask(Video currentVideo) {
        if (currentVideo == null) {
            Map<String, Object> emptyResult = new HashMap<>();
            emptyResult.put("processing", false);
            emptyResult.put("progress", 0);
            return emptyResult;
        }

        Map<String, Object> taskInfo = new HashMap<>();
        taskInfo.put("processing", true);
        taskInfo.put("videoId", currentVideo.getId());
        taskInfo.put("manuscriptId", currentVideo.getManuscriptId());
        taskInfo.put("videoTitle", currentVideo.getTitle());
        taskInfo.put("status", currentVideo.getProcessStatus());
        taskInfo.put("statusText", getStatusText(currentVideo.getProcessStatus()));
        taskInfo.put("progress", currentVideo.getProcessProgress() == null ? 0 : currentVideo.getProcessProgress());
        taskInfo.put("stage", currentVideo.getProcessStage());
        taskInfo.put("stageText", getStageText(currentVideo));
        taskInfo.put("error", currentVideo.getProcessError());
        return taskInfo;
    }

    private Map<String, Object> buildStatistics(List<Video> allVideos) {
        Map<String, Object> stats = new HashMap<>();
        int pending = 0;
        int transcoding = 0;
        int audioExtracting = 0;
        int subtitleGenerating = 0;
        int aiSummarizing = 0;
        int completed = 0;
        int failed = 0;

        for (Video video : allVideos) {
            Integer status = video.getProcessStatus();
            if (status == null) {
                status = Video.PROCESS_STATUS_PENDING;
            }

            switch (status) {
                case Video.PROCESS_STATUS_PENDING:
                    pending++;
                    break;
                case Video.PROCESS_STATUS_TRANSCODING:
                case Video.PROCESS_STATUS_TRANSCODE_SUCCESS:
                    transcoding++;
                    break;
                case Video.PROCESS_STATUS_AUDIO_EXTRACTING:
                case Video.PROCESS_STATUS_AUDIO_SUCCESS:
                    audioExtracting++;
                    break;
                case Video.PROCESS_STATUS_SUBTITLE_GENERATING:
                case Video.PROCESS_STATUS_SUBTITLE_SUCCESS:
                    subtitleGenerating++;
                    break;
                case Video.PROCESS_STATUS_AI_SUMMARIZING:
                case Video.PROCESS_STATUS_AI_SUCCESS:
                    aiSummarizing++;
                    break;
                case Video.PROCESS_STATUS_COMPLETED:
                    completed++;
                    break;
                case Video.PROCESS_STATUS_TRANSCODE_FAILED:
                case Video.PROCESS_STATUS_AUDIO_FAILED:
                case Video.PROCESS_STATUS_SUBTITLE_FAILED:
                case Video.PROCESS_STATUS_AI_FAILED:
                    failed++;
                    break;
                default:
                    break;
            }
        }

        stats.put("pending", pending);
        stats.put("transcoding", transcoding);
        stats.put("audioExtracting", audioExtracting);
        stats.put("subtitleGenerating", subtitleGenerating);
        stats.put("aiSummarizing", aiSummarizing);
        stats.put("completed", completed);
        stats.put("failed", failed);
        stats.put("total", allVideos.size());
        return stats;
    }

    private String getStageText(Video video) {
        if (video == null) {
            return "";
        }
        if (video.getProcessError() != null && !video.getProcessError().isEmpty()) {
            return video.getProcessError();
        }
        if (video.getProcessStage() == null || video.getProcessStage().isEmpty()) {
            return getStatusText(video.getProcessStatus());
        }
        return video.getProcessStage();
    }

    private boolean isProcessingStatus(Integer status) {
        return status == Video.PROCESS_STATUS_TRANSCODING
                || status == Video.PROCESS_STATUS_AUDIO_EXTRACTING
                || status == Video.PROCESS_STATUS_SUBTITLE_GENERATING
                || status == Video.PROCESS_STATUS_AI_SUMMARIZING;
    }

    private boolean isFailedStatus(Integer status) {
        return status == Video.PROCESS_STATUS_TRANSCODE_FAILED
                || status == Video.PROCESS_STATUS_AUDIO_FAILED
                || status == Video.PROCESS_STATUS_SUBTITLE_FAILED
                || status == Video.PROCESS_STATUS_AI_FAILED;
    }

    private String getStatusText(Integer status) {
        if (status == null) {
            return "未知";
        }
        switch (status) {
            case Video.PROCESS_STATUS_PENDING: return "待处理";
            case Video.PROCESS_STATUS_TRANSCODING: return "视频转码中";
            case Video.PROCESS_STATUS_TRANSCODE_FAILED: return "转码失败";
            case Video.PROCESS_STATUS_TRANSCODE_SUCCESS: return "转码成功";
            case Video.PROCESS_STATUS_AUDIO_EXTRACTING: return "音频提取中";
            case Video.PROCESS_STATUS_AUDIO_FAILED: return "音频提取失败";
            case Video.PROCESS_STATUS_AUDIO_SUCCESS: return "音频提取成功";
            case Video.PROCESS_STATUS_SUBTITLE_GENERATING: return "字幕生成中";
            case Video.PROCESS_STATUS_SUBTITLE_FAILED: return "字幕生成失败";
            case Video.PROCESS_STATUS_SUBTITLE_SUCCESS: return "字幕生成成功";
            case Video.PROCESS_STATUS_AI_SUMMARIZING: return "AI总结中";
            case Video.PROCESS_STATUS_AI_FAILED: return "AI总结失败";
            case Video.PROCESS_STATUS_AI_SUCCESS: return "AI总结成功";
            case Video.PROCESS_STATUS_COMPLETED: return "处理完成";
            default: return "状态: " + status;
        }
    }
}
