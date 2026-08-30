package com.mybilibili.admin.controller;

import com.mybilibili.admin.mapper.ManuscriptMapper;
import com.mybilibili.admin.mapper.VideoMapper;
import com.mybilibili.admin.mapper.UserMapper;
import com.mybilibili.admin.service.AiSummaryService;
import com.mybilibili.admin.service.VideoProcessService;
import com.mybilibili.admin.utils.SubtitleTextUtils;
import com.mybilibili.admin.utils.UploadFilePathUtils;
import com.mybilibili.common.entity.Manuscript;
import com.mybilibili.common.entity.Video;
import com.mybilibili.common.entity.User;
import com.mybilibili.common.service.VideoIndexService;
import com.mybilibili.common.vo.Result;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.web.bind.annotation.*;

import java.io.File;
import java.util.Date;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.ArrayList;
import java.util.stream.Collectors;

@RestController
@RequestMapping("/manuscript")
@Tag(name = "管理员稿件管理", description = "稿件审核、上架、下架等功能")
public class AdminManuscriptController {

    @Autowired
    private ManuscriptMapper manuscriptMapper;

    @Autowired
    private VideoMapper videoMapper;

    @Autowired
    private UserMapper userMapper;

    @Autowired
    private StringRedisTemplate redisTemplate;

    @Autowired
    private VideoProcessService videoProcessService;

    @Autowired(required = false)
    private VideoIndexService videoIndexService;

    @Autowired
    private AiSummaryService aiSummaryService;

    @Autowired
    private UploadFilePathUtils uploadFilePathUtils;

    // Redis Key 前缀
    private static final String KEY_CURRENT = "manuscript:process:%s:current";
    private static final String KEY_QUEUE = "manuscript:process:%s:queue";

    @Operation(summary = "获取待审核列表", description = "获取所有待审核的稿件列表")
    @GetMapping("/pending")
    public Result<List<Manuscript>> getPendingManuscripts() {
        List<Manuscript> manuscripts = manuscriptMapper.selectByStatus(Manuscript.STATUS_PENDING_REVIEW);
        return Result.success(manuscripts);
    }

    @Operation(summary = "审核通过", description = "审核通过稿件，状态改为处理中")
    @PostMapping("/approve/{manuscriptId}")
    public Result<Map<String, Object>> approveManuscript(
            @PathVariable Integer manuscriptId,
            @RequestParam(required = false) String reason,
            @RequestParam Integer reviewerId) {

        Manuscript manuscript = manuscriptMapper.selectById(manuscriptId);
        if (manuscript == null) {
            return Result.error("稿件不存在");
        }

        if (manuscript.getStatus() != Manuscript.STATUS_PENDING_REVIEW) {
            return Result.error("稿件状态不是待审核");
        }

        Manuscript updateManuscript = new Manuscript();
        updateManuscript.setId(manuscriptId);
        updateManuscript.setStatus(Manuscript.STATUS_PROCESSING);
        updateManuscript.setReviewStatus(Manuscript.REVIEW_STATUS_APPROVED);
        updateManuscript.setReviewReason(reason);
        updateManuscript.setReviewTime(new Date());
        updateManuscript.setReviewerId(reviewerId);

        manuscriptMapper.update(updateManuscript);

        List<Video> videos = videoMapper.selectByManuscriptId(manuscriptId);
        for (Video video : videos) {
            Video updateVideo = new Video();
            updateVideo.setId(video.getId());
            updateVideo.setStatus(Video.STATUS_PROCESSING);
            updateVideo.setReviewStatus(Video.REVIEW_STATUS_APPROVED);
            updateVideo.setReviewReason(reason);
            updateVideo.setReviewTime(new Date());
            updateVideo.setReviewerId(reviewerId);
            updateVideo.setProcessStatus(Video.PROCESS_STATUS_PENDING);
            videoMapper.update(updateVideo);
        }

        // 初始化 Redis 队列
        initManuscriptQueue(manuscriptId, videos);

        Map<String, Object> data = new HashMap<>();
        data.put("manuscriptId", manuscriptId);
        data.put("status", Manuscript.STATUS_PROCESSING);
        data.put("message", "审核通过，定时任务将自动开始处理");

        return Result.success(data);
    }

    private void initManuscriptQueue(Integer manuscriptId, List<Video> videos) {
        String queueKey = String.format(KEY_QUEUE, manuscriptId);
        List<String> videoIds = videos.stream()
            .map(v -> String.valueOf(v.getId()))
            .collect(Collectors.toList());
        
        if (!videoIds.isEmpty()) {
            redisTemplate.opsForList().rightPushAll(queueKey, videoIds);
            System.out.println("稿件 " + manuscriptId + " 初始化 Redis 队列，共 " + videoIds.size() + " 个视频");
        }
    }

    @Operation(summary = "审核拒绝", description = "审核拒绝稿件")
    @PostMapping("/reject/{manuscriptId}")
    public Result<Map<String, Object>> rejectManuscript(
            @PathVariable Integer manuscriptId,
            @RequestParam String reason,
            @RequestParam Integer reviewerId) {

        Manuscript manuscript = manuscriptMapper.selectById(manuscriptId);
        if (manuscript == null) {
            return Result.error("稿件不存在");
        }

        if (manuscript.getStatus() != Manuscript.STATUS_PENDING_REVIEW) {
            return Result.error("稿件状态不是待审核");
        }

        Manuscript updateManuscript = new Manuscript();
        updateManuscript.setId(manuscriptId);
        updateManuscript.setStatus(Manuscript.STATUS_REJECTED);
        updateManuscript.setReviewStatus(Manuscript.REVIEW_STATUS_REJECTED);
        updateManuscript.setReviewReason(reason);
        updateManuscript.setReviewTime(new Date());
        updateManuscript.setReviewerId(reviewerId);

        manuscriptMapper.update(updateManuscript);

        List<Video> videos = videoMapper.selectByManuscriptId(manuscriptId);
        for (Video video : videos) {
            Video updateVideo = new Video();
            updateVideo.setId(video.getId());
            updateVideo.setStatus(Video.STATUS_REJECTED);
            updateVideo.setReviewStatus(Video.REVIEW_STATUS_REJECTED);
            updateVideo.setReviewReason(reason);
            updateVideo.setReviewTime(new Date());
            updateVideo.setReviewerId(reviewerId);
            videoMapper.update(updateVideo);
        }

        Map<String, Object> data = new HashMap<>();
        data.put("manuscriptId", manuscriptId);
        data.put("status", Manuscript.STATUS_REJECTED);
        data.put("message", "审核已拒绝");

        return Result.success(data);
    }

    @Operation(summary = "获取处理中列表", description = "获取所有处理中的稿件列表")
    @GetMapping("/processing")
    public Result<List<Manuscript>> getProcessingManuscripts() {
        List<Manuscript> manuscripts = manuscriptMapper.selectByStatus(Manuscript.STATUS_PROCESSING);
        return Result.success(manuscripts);
    }

    @Operation(summary = "获取待上架列表", description = "获取所有处理完成待上架的稿件列表")
    @GetMapping("/ready")
    public Result<List<Manuscript>> getReadyManuscripts() {
        List<Manuscript> manuscripts = manuscriptMapper.selectByStatus(Manuscript.STATUS_READY_TO_PUBLISH);
        return Result.success(manuscripts);
    }

    @Operation(summary = "上架稿件", description = "将稿件上架，用户可见")
    @PostMapping("/publish/{manuscriptId}")
    public Result<Map<String, Object>> publishManuscript(@PathVariable Integer manuscriptId) {
        Manuscript manuscript = manuscriptMapper.selectById(manuscriptId);
        if (manuscript == null) {
            return Result.error("稿件不存在");
        }

        if (manuscript.getStatus() != Manuscript.STATUS_READY_TO_PUBLISH && manuscript.getStatus() != Manuscript.STATUS_UNPUBLISHED) {
            return Result.error("稿件状态不允许上架");
        }

        Manuscript updateManuscript = new Manuscript();
        updateManuscript.setId(manuscriptId);
        updateManuscript.setStatus(Manuscript.STATUS_PUBLISHED);

        manuscriptMapper.update(updateManuscript);

        List<Video> videos = videoMapper.selectByManuscriptId(manuscriptId);
        for (Video video : videos) {
            Video updateVideo = new Video();
            updateVideo.setId(video.getId());
            updateVideo.setStatus(Video.STATUS_PUBLISHED);
            videoMapper.update(updateVideo);

            // 更新视频索引
            if (videoIndexService != null) {
                video.setStatus(Video.STATUS_PUBLISHED);
                videoIndexService.indexVideo(video);
            }
        }

        Map<String, Object> data = new HashMap<>();
        data.put("manuscriptId", manuscriptId);
        data.put("status", Manuscript.STATUS_PUBLISHED);
        data.put("message", "稿件已上架");

        return Result.success(data);
    }

    @Operation(summary = "下架稿件", description = "将稿件下架")
    @PostMapping("/unpublish/{manuscriptId}")
    public Result<Map<String, Object>> unpublishManuscript(@PathVariable Integer manuscriptId) {
        try {
            Manuscript manuscript = manuscriptMapper.selectById(manuscriptId);
            if (manuscript == null) {
                return Result.error("稿件不存在");
            }

            if (manuscript.getStatus() != Manuscript.STATUS_PUBLISHED) {
                return Result.error("稿件状态不是已上架");
            }

            manuscriptMapper.updateStatus(manuscriptId, Manuscript.STATUS_UNPUBLISHED);

            List<Video> videos = videoMapper.selectByManuscriptId(manuscriptId);
            for (Video video : videos) {
                videoMapper.updateStatus(video.getId(), Video.STATUS_UNPUBLISHED);

                // 删除视频索引
                try {
                    if (videoIndexService != null) {
                        videoIndexService.deleteVideo(video.getId());
                    }
                } catch (Exception e) {
                    // 索引删除失败不影响下架操作，记录日志即可
                    System.err.println("删除视频索引失败: " + video.getId() + ", " + e.getMessage());
                }
            }

            Map<String, Object> data = new HashMap<>();
            data.put("manuscriptId", manuscriptId);
            data.put("status", Manuscript.STATUS_UNPUBLISHED);
            data.put("message", "稿件已下架");

            return Result.success(data);
        } catch (Exception e) {
            e.printStackTrace();
            return Result.error("下架失败: " + e.getMessage());
        }
    }

    @Operation(summary = "获取所有稿件", description = "获取所有稿件列表（管理员）")
    @GetMapping("/all")
    public Result<List<Map<String, Object>>> getAllManuscripts() {
        List<Manuscript> manuscripts = manuscriptMapper.selectAll();
        List<Map<String, Object>> result = new ArrayList<>();
        
        for (Manuscript manuscript : manuscripts) {
            Map<String, Object> manuscriptMap = new HashMap<>();
            manuscriptMap.put("id", manuscript.getId());
            manuscriptMap.put("title", manuscript.getTitle());
            manuscriptMap.put("description", manuscript.getDescription());
            manuscriptMap.put("coverUrl", manuscript.getCoverUrl());
            manuscriptMap.put("userId", manuscript.getUserId());
            manuscriptMap.put("status", manuscript.getStatus());
            manuscriptMap.put("reviewStatus", manuscript.getReviewStatus());
            manuscriptMap.put("reviewReason", manuscript.getReviewReason());
            manuscriptMap.put("reviewTime", manuscript.getReviewTime());
            manuscriptMap.put("reviewerId", manuscript.getReviewerId());
            manuscriptMap.put("viewCount", manuscript.getViewCount());
            manuscriptMap.put("likeCount", manuscript.getLikeCount());
            manuscriptMap.put("collectCount", manuscript.getCollectCount());
            manuscriptMap.put("durationSeconds", manuscript.getDurationSeconds());
            manuscriptMap.put("uploadTime", manuscript.getUploadTime());
            manuscriptMap.put("updatedAt", manuscript.getUpdatedAt());
            
            // 获取视频列表
            List<Video> videos = videoMapper.selectByManuscriptId(manuscript.getId());
            manuscriptMap.put("videos", videos);
            
            result.add(manuscriptMap);
        }
        
        return Result.success(result);
    }

    @Operation(summary = "获取稿件视频列表", description = "获取指定稿件下的所有视频列表")
    @GetMapping("/{manuscriptId}/videos")
    public Result<List<Video>> getManuscriptVideos(@PathVariable Integer manuscriptId) {
        Manuscript manuscript = manuscriptMapper.selectById(manuscriptId);
        if (manuscript == null) {
            return Result.error("稿件不存在");
        }

        List<Video> videos = videoMapper.selectByManuscriptId(manuscriptId);
        return Result.success(videos);
    }

    @Operation(summary = "获取稿件详情", description = "获取稿件详情，包含视频列表和用户信息")
    @GetMapping("/{manuscriptId}")
    public Result<Map<String, Object>> getManuscriptDetail(@PathVariable Integer manuscriptId) {
        Manuscript manuscript = manuscriptMapper.selectById(manuscriptId);
        if (manuscript == null) {
            return Result.error("稿件不存在");
        }

        List<Video> videos = videoMapper.selectByManuscriptId(manuscriptId);
        
        User user = userMapper.selectById(manuscript.getUserId());

        Map<String, Object> result = new HashMap<>();
        result.put("manuscript", manuscript);
        result.put("videos", videos);
        result.put("user", user);

        return Result.success(result);
    }

    @Operation(summary = "获取稿件统计", description = "获取各状态的稿件数量统计")
    @GetMapping("/statistics")
    public Result<Map<String, Object>> getManuscriptStatistics() {
        Map<String, Object> result = new HashMap<>();
        
        result.put("total", manuscriptMapper.countAll());
        result.put("pending", manuscriptMapper.countByStatus(Manuscript.STATUS_PENDING_REVIEW));
        result.put("processing", manuscriptMapper.countByStatus(Manuscript.STATUS_PROCESSING));
        result.put("ready", manuscriptMapper.countByStatus(Manuscript.STATUS_READY_TO_PUBLISH));
        result.put("published", manuscriptMapper.countByStatus(Manuscript.STATUS_PUBLISHED));
        result.put("rejected", manuscriptMapper.countByStatus(Manuscript.STATUS_REJECTED));
        result.put("failed", manuscriptMapper.countByStatus(Manuscript.STATUS_PROCESS_FAILED));
        result.put("unpublished", manuscriptMapper.countByStatus(Manuscript.STATUS_UNPUBLISHED));

        return Result.success(result);
    }

    @Operation(summary = "重试处理稿件", description = "重试处理失败的稿件")
    @PostMapping("/retry/{manuscriptId}")
    public Result<Map<String, Object>> retryManuscript(@PathVariable Integer manuscriptId) {
        Manuscript manuscript = manuscriptMapper.selectById(manuscriptId);
        if (manuscript == null) {
            return Result.error("稿件不存在");
        }

        if (manuscript.getStatus() != Manuscript.STATUS_PROCESS_FAILED) {
            return Result.error("稿件状态不是处理失败");
        }

        // 清理 Redis 数据
        clearManuscriptProcessData(manuscriptId);

        Manuscript updateManuscript = new Manuscript();
        updateManuscript.setId(manuscriptId);
        updateManuscript.setStatus(Manuscript.STATUS_PROCESSING);

        manuscriptMapper.update(updateManuscript);

        List<Video> videos = videoMapper.selectByManuscriptId(manuscriptId);
        for (Video video : videos) {
            if (video.getProcessStatus() != null && video.getProcessStatus() >= Video.PROCESS_STATUS_TRANSCODE_FAILED 
                && video.getProcessStatus() <= Video.PROCESS_STATUS_AI_FAILED) {
                Video updateVideo = new Video();
                updateVideo.setId(video.getId());
                updateVideo.setProcessStatus(Video.PROCESS_STATUS_PENDING);
                updateVideo.setProcessError(null);
                videoMapper.update(updateVideo);
            }
        }

        // 重新初始化 Redis 队列
        initManuscriptQueue(manuscriptId, videos);

        Map<String, Object> data = new HashMap<>();
        data.put("manuscriptId", manuscriptId);
        data.put("status", Manuscript.STATUS_PROCESSING);
        data.put("message", "已重新开始处理");

        return Result.success(data);
    }

    private void clearManuscriptProcessData(Integer manuscriptId) {
        String currentKey = String.format(KEY_CURRENT, manuscriptId);
        String queueKey = String.format(KEY_QUEUE, manuscriptId);
        redisTemplate.delete(currentKey);
        redisTemplate.delete(queueKey);
        System.out.println("稿件 " + manuscriptId + " 清理 Redis 数据");
    }

    // ==================== 手动处理流程 API ====================

    @Operation(summary = "手动开始转码", description = "手动触发视频转码")
    @PostMapping("/transcode/{videoId}")
    public Result<Map<String, Object>> manualTranscode(@PathVariable Integer videoId) {
        Video video = videoMapper.selectById(videoId);
        if (video == null) {
            return Result.error("视频不存在");
        }

        // 检查是否正在处理中（防止重复提交）
        if (video.getProcessStatus() != null && 
            (video.getProcessStatus() == Video.PROCESS_STATUS_TRANSCODING ||
             video.getProcessStatus() == Video.PROCESS_STATUS_AUDIO_EXTRACTING ||
             video.getProcessStatus() == Video.PROCESS_STATUS_SUBTITLE_GENERATING ||
             video.getProcessStatus() == Video.PROCESS_STATUS_AI_SUMMARIZING)) {
            return Result.error("视频正在处理中，请等待完成后再试");
        }

        // 异步执行转码
        new Thread(() -> {
            videoProcessService.transcodeVideo(videoId);
        }).start();

        Map<String, Object> data = new HashMap<>();
        data.put("videoId", videoId);
        data.put("status", Video.PROCESS_STATUS_TRANSCODING);
        data.put("message", "已开始转码，请稍后刷新查看状态");

        return Result.success(data);
    }

    @Operation(summary = "手动提取音频", description = "手动触发音频提取")
    @PostMapping("/extract-audio/{videoId}")
    public Result<Map<String, Object>> manualExtractAudio(@PathVariable Integer videoId) {
        Video video = videoMapper.selectById(videoId);
        if (video == null) {
            return Result.error("视频不存在");
        }

        // 检查是否正在处理中（防止重复提交）
        if (video.getProcessStatus() != null && 
            (video.getProcessStatus() == Video.PROCESS_STATUS_TRANSCODING ||
             video.getProcessStatus() == Video.PROCESS_STATUS_AUDIO_EXTRACTING ||
             video.getProcessStatus() == Video.PROCESS_STATUS_SUBTITLE_GENERATING ||
             video.getProcessStatus() == Video.PROCESS_STATUS_AI_SUMMARIZING)) {
            return Result.error("视频正在处理中，请等待完成后再试");
        }

        new Thread(() -> {
            videoProcessService.extractAudio(videoId);
        }).start();

        Map<String, Object> data = new HashMap<>();
        data.put("videoId", videoId);
        data.put("status", Video.PROCESS_STATUS_AUDIO_EXTRACTING);
        data.put("message", "已开始提取音频");

        return Result.success(data);
    }

    @Operation(summary = "手动生成字幕", description = "手动触发字幕生成")
    @PostMapping("/generate-subtitle/{videoId}")
    public Result<Map<String, Object>> manualGenerateSubtitle(@PathVariable Integer videoId) {
        Video video = videoMapper.selectById(videoId);
        if (video == null) {
            return Result.error("视频不存在");
        }

        // 检查是否正在处理中（防止重复提交）
        if (video.getProcessStatus() != null && 
            (video.getProcessStatus() == Video.PROCESS_STATUS_TRANSCODING ||
             video.getProcessStatus() == Video.PROCESS_STATUS_AUDIO_EXTRACTING ||
             video.getProcessStatus() == Video.PROCESS_STATUS_SUBTITLE_GENERATING ||
             video.getProcessStatus() == Video.PROCESS_STATUS_AI_SUMMARIZING)) {
            return Result.error("视频正在处理中，请等待完成后再试");
        }

        new Thread(() -> {
            videoProcessService.generateSubtitle(videoId);
        }).start();

        Map<String, Object> data = new HashMap<>();
        data.put("videoId", videoId);
        data.put("status", Video.PROCESS_STATUS_SUBTITLE_GENERATING);
        data.put("message", "已开始生成字幕");

        return Result.success(data);
    }

    @Operation(summary = "手动AI总结", description = "手动触发AI总结")
    @PostMapping("/ai-summary/{videoId}")
    public Result<Map<String, Object>> manualAiSummary(@PathVariable Integer videoId) {
        Video video = videoMapper.selectById(videoId);
        if (video == null) {
            return Result.error("视频不存在");
        }

        // 检查是否正在处理中（防止重复提交）
        if (video.getProcessStatus() != null && 
            (video.getProcessStatus() == Video.PROCESS_STATUS_TRANSCODING ||
             video.getProcessStatus() == Video.PROCESS_STATUS_AUDIO_EXTRACTING ||
             video.getProcessStatus() == Video.PROCESS_STATUS_SUBTITLE_GENERATING ||
             video.getProcessStatus() == Video.PROCESS_STATUS_AI_SUMMARIZING)) {
            return Result.error("视频正在处理中，请等待完成后再试");
        }

        new Thread(() -> {
            videoProcessService.aiSummary(videoId);
        }).start();

        Map<String, Object> data = new HashMap<>();
        data.put("videoId", videoId);
        data.put("status", Video.PROCESS_STATUS_AI_SUMMARIZING);
        data.put("message", "已开始AI总结");

        return Result.success(data);
    }

    @Operation(summary = "一键处理视频", description = "一键执行所有处理步骤（转码->音频->字幕->AI）")
    @PostMapping("/process-all/{videoId}")
    public Result<Map<String, Object>> manualProcessAll(@PathVariable Integer videoId) {
        Video video = videoMapper.selectById(videoId);
        if (video == null) {
            return Result.error("视频不存在");
        }

        new Thread(() -> {
            videoProcessService.processAll(videoId);
        }).start();

        Map<String, Object> data = new HashMap<>();
        data.put("videoId", videoId);
        data.put("message", "已开始全流程处理");

        return Result.success(data);
    }

    @Operation(summary = "重置视频状态", description = "将视频状态重置为待处理")
    @PostMapping("/reset/{videoId}")
    public Result<Map<String, Object>> resetVideoStatus(@PathVariable Integer videoId) {
        Video video = videoMapper.selectById(videoId);
        if (video == null) {
            return Result.error("视频不存在");
        }

        Video updateVideo = new Video();
        updateVideo.setId(videoId);
        updateVideo.setProcessStatus(Video.PROCESS_STATUS_PENDING);
        updateVideo.setProcessError(null);
        videoMapper.update(updateVideo);

        Map<String, Object> data = new HashMap<>();
        data.put("videoId", videoId);
        data.put("status", Video.PROCESS_STATUS_PENDING);
        data.put("message", "视频状态已重置");

        return Result.success(data);
    }

    // ==================== 视频处理查询接口 ====================

    @Operation(summary = "获取视频处理状态", description = "获取视频的详细处理状态")
    @GetMapping("/video-status/{videoId}")
    public Result<Map<String, Object>> getVideoProcessStatus(@PathVariable Integer videoId) {
        Video video = videoMapper.selectById(videoId);
        if (video == null) {
            return Result.error("视频不存在");
        }

        Map<String, Object> data = new HashMap<>();
        data.put("videoId", videoId);
        data.put("processStatus", video.getProcessStatus());
        data.put("processStatusText", getProcessStatusText(video.getProcessStatus()));
        data.put("processError", video.getProcessError());
        data.put("hasSubtitle", video.getHasSubtitle());
        data.put("hasSummary", video.getHasSummary());
        data.put("playUrl", video.getPlayUrl());
        data.put("sourceVideoUrl", video.getSourceVideoUrl());

        return Result.success(data);
    }

    @Operation(summary = "获取源视频预览URL", description = "获取视频源文件URL（用于审核预览）")
    @GetMapping("/video-source/{videoId}")
    public Result<Map<String, Object>> getVideoSourceUrl(@PathVariable Integer videoId) {
        Video video = videoMapper.selectById(videoId);
        if (video == null) {
            return Result.error("视频不存在");
        }

        // 动态生成源视频URL（避免数据库中保存的URL不正确）
        String sourceUrl = "/uploads/manuscripts/" + video.getManuscriptId() + "/videos/" + videoId + "/source/video.mp4";

        Map<String, Object> data = new HashMap<>();
        data.put("videoId", videoId);
        data.put("sourceUrl", sourceUrl);
        data.put("title", video.getTitle());
        data.put("durationSeconds", video.getDurationSeconds());

        return Result.success(data);
    }

    private String getProcessStatusText(Integer status) {
        if (status == null) return "未知";
        switch (status) {
            case Video.PROCESS_STATUS_PENDING: return "待处理";
            case Video.PROCESS_STATUS_TRANSCODING: return "视频转码中";
            case Video.PROCESS_STATUS_TRANSCODE_SUCCESS: return "转码成功";
            case Video.PROCESS_STATUS_TRANSCODE_FAILED: return "转码失败";
            case Video.PROCESS_STATUS_AUDIO_EXTRACTING: return "音频提取中";
            case Video.PROCESS_STATUS_AUDIO_SUCCESS: return "音频提取成功";
            case Video.PROCESS_STATUS_AUDIO_FAILED: return "音频提取失败";
            case Video.PROCESS_STATUS_SUBTITLE_GENERATING: return "字幕生成中";
            case Video.PROCESS_STATUS_SUBTITLE_SUCCESS: return "字幕生成成功";
            case Video.PROCESS_STATUS_SUBTITLE_FAILED: return "字幕生成失败";
            case Video.PROCESS_STATUS_AI_SUMMARIZING: return "AI总结中";
            case Video.PROCESS_STATUS_AI_SUCCESS: return "AI总结成功";
            case Video.PROCESS_STATUS_AI_FAILED: return "AI总结失败";
            case Video.PROCESS_STATUS_COMPLETED: return "处理完成";
            default: return "未知";
        }
    }

    // ==================== AI摘要测试接口 ====================

    @Operation(summary = "测试AI API连接", description = "测试DeepSeek API是否配置正确")
    @PostMapping("/test-ai-api")
    public Result<Map<String, Object>> testAiApi(@RequestBody(required = false) Map<String, String> request) {
        String testText = request != null ? request.get("text") : null;

        AiSummaryService.TestResult testResult = aiSummaryService.testApiConnection(testText);

        Map<String, Object> data = new HashMap<>();
        data.put("success", testResult.isSuccess());
        data.put("message", testResult.getMessage());
        data.put("response", testResult.getResponse());
        data.put("responseTime", testResult.getResponseTime() + "ms");

        if (testResult.isSuccess()) {
            return Result.success(data);
        } else {
            return Result.error(500, testResult.getMessage());
        }
    }

    @Operation(summary = "测试AI摘要生成", description = "使用指定视频的字幕内容测试AI摘要生成")
    @PostMapping("/test-ai-summary/{videoId}")
    public Result<Map<String, Object>> testAiSummary(@PathVariable Integer videoId) {
        Video video = videoMapper.selectById(videoId);
        if (video == null) {
            return Result.error("视频不存在");
        }

        Integer manuscriptId = video.getManuscriptId();
        if (manuscriptId == null) {
            return Result.error("视频未关联稿件");
        }

        try {
            // 读取字幕内容
            String subtitlePath = uploadFilePathUtils.getChineseSubtitlePath(manuscriptId, videoId);
            String subtitlePlainText = "";

            File subtitleFile = new File(subtitlePath);
            if (subtitleFile.exists()) {
                subtitlePlainText = SubtitleTextUtils.extractPlainText(subtitlePath);
            }

            // 获取稿件信息
            Manuscript manuscript = manuscriptMapper.selectById(manuscriptId);
            String videoTitle = video.getTitle();
            String videoDescription = video.getDescription();

            if (manuscript != null) {
                if (videoTitle == null || videoTitle.isEmpty()) {
                    videoTitle = manuscript.getTitle();
                }
                if (videoDescription == null || videoDescription.isEmpty()) {
                    videoDescription = manuscript.getDescription();
                }
            }

            // 调用AI服务生成摘要
            long startTime = System.currentTimeMillis();
            String summary = aiSummaryService.generateSummary(subtitlePlainText, videoTitle, videoDescription);
            long responseTime = System.currentTimeMillis() - startTime;

            Map<String, Object> data = new HashMap<>();
            data.put("videoId", videoId);
            data.put("videoTitle", videoTitle);
            data.put("subtitleLength", subtitlePlainText.length());
            data.put("subtitleTokenEstimate", SubtitleTextUtils.estimateTokenCount(subtitlePlainText));
            data.put("summaryLength", summary.length());
            data.put("responseTime", responseTime + "ms");
            data.put("summary", summary);

            return Result.success(data);

        } catch (Exception e) {
            return Result.error("测试失败: " + e.getMessage());
        }
    }
}
