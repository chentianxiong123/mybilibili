package com.mybilibili.ai.service.impl;

import com.mybilibili.ai.mapper.VideoMapper;
import com.mybilibili.ai.pipeline.PipelineTask;
import com.mybilibili.ai.pipeline.VideoPipelineQueueManager;
import com.mybilibili.ai.service.*;
import com.mybilibili.ai.utils.SubtitleTextUtils;
import com.mybilibili.ai.websocket.VideoProcessWebSocketHandler;
import com.mybilibili.common.entity.Video;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;

import javax.annotation.PostConstruct;
import java.util.List;
import java.util.Map;

/**
 * 视频全流程处理服务实现
 * 实现自动按顺序执行视频的所有处理步骤（转码→音频→字幕→AI）
 */
@Service
public class VideoPipelineServiceImpl implements VideoPipelineService {

    private static final Logger log = LoggerFactory.getLogger(VideoPipelineServiceImpl.class);

    @Autowired
    private VideoPipelineQueueManager queueManager;

    @Autowired
    private VideoTranscodeService videoTranscodeService;

    @Autowired
    private AudioExtractService audioExtractService;

    @Autowired
    private AiSubtitleService aiSubtitleService;

    @Autowired
    private AiSummaryService aiSummaryService;

    @Autowired
    private VideoProgressSseService progressSseService;

    @Autowired
    private VideoMapper videoMapper;

    @Autowired
    private StringRedisTemplate redisTemplate;

    /**
     * 初始化：设置任务处理器并启动队列管理器
     */
    @PostConstruct
    public void init() {
        queueManager.setTaskProcessor(this::processTask);
        queueManager.start();
        log.info("[全流程服务] 初始化完成");
    }

    @Override
    public boolean submitPipelineTask(Integer manuscriptId, Integer videoId, Integer uploaderId) {
        // 获取视频信息
        Video video = videoMapper.selectById(videoId);
        String videoTitle = video != null ? video.getTitle() : "未知视频";

        // 创建任务
        PipelineTask task = PipelineTask.create(manuscriptId, videoId, uploaderId, videoTitle);

        // 设置视频标题到进度服务
        progressSseService.setVideoTitle(videoId, videoTitle);

        // 提交到队列
        boolean submitted = queueManager.submitTask(task);

        if (submitted) {
            log.info("[全流程服务] 任务已提交: manuscriptId={}, videoId={}, title={}",
                    manuscriptId, videoId, videoTitle);
            // 更新视频状态为处理中
            updateVideoStatus(videoId, Video.PROCESS_STATUS_TRANSCODING);
        }

        return submitted;
    }

    @Override
    public boolean cancelTask(Integer manuscriptId, Integer videoId) {
        String taskKey = manuscriptId + "-" + videoId;
        return queueManager.cancelTask(taskKey);
    }

    @Override
    public PipelineTask getTaskStatus(Integer manuscriptId, Integer videoId) {
        String taskKey = manuscriptId + "-" + videoId;
        return queueManager.getTaskStatus(taskKey);
    }

    @Override
    public Map<String, Object> getStatistics() {
        return queueManager.getStatistics();
    }

    @Override
    public List<PipelineTask> getQueuedTasks() {
        return queueManager.getQueuedTasks();
    }

    @Override
    public List<PipelineTask> getProcessingTasks() {
        return queueManager.getProcessingTasks();
    }

    @Override
    public List<PipelineTask> getCompletedTasks() {
        return queueManager.getCompletedTasks();
    }

    @Override
    public PipelineTask getCurrentTask() {
        return queueManager.getCurrentTask();
    }

    @Override
    public void clearQueue() {
        queueManager.clearQueue();
    }

    @Override
    public boolean retryTask(Integer manuscriptId, Integer videoId) {
        // 获取之前的任务状态
        PipelineTask oldTask = getTaskStatus(manuscriptId, videoId);

        if (oldTask == null) {
            log.warn("[全流程服务] 未找到任务: manuscriptId={}, videoId={}", manuscriptId, videoId);
            return false;
        }

        // 只有失败的任务才能重试
        if (oldTask.getStatus() != PipelineTask.TaskStatus.FAILED) {
            log.warn("[全流程服务] 任务状态不允许重试: {}", oldTask.getStatus());
            return false;
        }

        // 重新提交任务
        return submitPipelineTask(manuscriptId, videoId, oldTask.getUploaderId());
    }

    /**
     * 处理任务（核心处理逻辑）
     */
    private void processTask(PipelineTask task) {
        Integer videoId = task.getVideoId();
        Integer manuscriptId = task.getManuscriptId();
        String videoTitle = task.getVideoTitle();

        log.info("[全流程服务] 开始处理任务: {}", task.getTaskKey());

        try {
            // 广播开始
            broadcastProgress(videoId, manuscriptId, videoTitle, "STARTING", 0, null);

            // 按顺序执行每个步骤
            while (!task.isAllStepsCompleted() && task.getStatus() == PipelineTask.TaskStatus.RUNNING) {
                PipelineTask.ProcessStep currentStep = task.getCurrentStep();

                log.info("[全流程服务] 执行步骤: {} - {}", task.getTaskKey(), currentStep.getDescription());

                boolean success = executeStep(task, currentStep);

                if (!success) {
                    // 步骤失败，停止后续步骤
                    log.error("[全流程服务] 步骤失败: {} - {}", task.getTaskKey(), currentStep.getDescription());
                    task.markFailed(currentStep.getCode(), currentStep.getDescription() + "失败");
                    broadcastError(videoId, manuscriptId, videoTitle, currentStep.getCode(), currentStep.getDescription() + "失败");
                    return;
                }

                // 步骤成功，前进到下一步
                boolean hasMoreSteps = task.advanceToNextStep();

                if (!hasMoreSteps) {
                    // 所有步骤完成
                    task.markCompleted();
                    log.info("[全流程服务] 任务完成: {}", task.getTaskKey());
                    broadcastProgress(videoId, manuscriptId, videoTitle, "COMPLETED", 100, Video.PROCESS_STATUS_COMPLETED);
                    VideoProcessWebSocketHandler.broadcastComplete(videoId, manuscriptId, videoTitle);
                    progressSseService.complete(videoId);
                }
            }

        } catch (Exception e) {
            log.error("[全流程服务] 任务处理异常: {}", task.getTaskKey(), e);
            task.markFailed("UNKNOWN", e.getMessage());
            broadcastError(videoId, manuscriptId, videoTitle, "UNKNOWN", e.getMessage());
        }
    }

    /**
     * 执行单个步骤
     */
    private boolean executeStep(PipelineTask task, PipelineTask.ProcessStep step) {
        Integer videoId = task.getVideoId();
        Integer manuscriptId = task.getManuscriptId();
        String videoTitle = task.getVideoTitle();

        switch (step) {
            case TRANSCODE:
                return executeTranscode(manuscriptId, videoId, videoTitle);
            case AUDIO_EXTRACT:
                return executeAudioExtract(manuscriptId, videoId, videoTitle);
            case SUBTITLE_GENERATE:
                return executeSubtitleGenerate(manuscriptId, videoId, videoTitle);
            case AI_SUMMARY:
                return executeAiSummary(manuscriptId, videoId, videoTitle);
            default:
                return false;
        }
    }

    /**
     * 执行转码步骤
     */
    private boolean executeTranscode(Integer manuscriptId, Integer videoId, String videoTitle) {
        log.info("[全流程服务] 开始转码: videoId={}", videoId);
        broadcastProgress(videoId, manuscriptId, videoTitle, "TRANSCODING", 0, Video.PROCESS_STATUS_TRANSCODING);
        updateVideoStatus(videoId, Video.PROCESS_STATUS_TRANSCODING);

        boolean success = videoTranscodeService.transcode(manuscriptId, videoId);

        if (success) {
            broadcastProgress(videoId, manuscriptId, videoTitle, "TRANSCODE_SUCCESS", 100, Video.PROCESS_STATUS_TRANSCODE_SUCCESS);
            updateVideoStatus(videoId, Video.PROCESS_STATUS_TRANSCODE_SUCCESS);
            log.info("[全流程服务] 转码完成: videoId={}", videoId);
        } else {
            broadcastError(videoId, manuscriptId, videoTitle, "TRANSCODE", "转码失败");
            updateVideoStatus(videoId, Video.PROCESS_STATUS_TRANSCODE_FAILED);
            log.error("[全流程服务] 转码失败: videoId={}", videoId);
        }

        return success;
    }

    /**
     * 执行音频提取步骤
     */
    private boolean executeAudioExtract(Integer manuscriptId, Integer videoId, String videoTitle) {
        log.info("[全流程服务] 开始音频提取: videoId={}", videoId);
        broadcastProgress(videoId, manuscriptId, videoTitle, "AUDIO_EXTRACTING", 0, Video.PROCESS_STATUS_AUDIO_EXTRACTING);
        updateVideoStatus(videoId, Video.PROCESS_STATUS_AUDIO_EXTRACTING);

        boolean success = audioExtractService.extractAudio(manuscriptId, videoId);

        if (success) {
            broadcastProgress(videoId, manuscriptId, videoTitle, "AUDIO_SUCCESS", 100, Video.PROCESS_STATUS_AUDIO_SUCCESS);
            updateVideoStatus(videoId, Video.PROCESS_STATUS_AUDIO_SUCCESS);
            log.info("[全流程服务] 音频提取完成: videoId={}", videoId);
        } else {
            broadcastError(videoId, manuscriptId, videoTitle, "EXTRACT_AUDIO", "音频提取失败");
            updateVideoStatus(videoId, Video.PROCESS_STATUS_AUDIO_FAILED);
            log.error("[全流程服务] 音频提取失败: videoId={}", videoId);
        }

        return success;
    }

    /**
     * 执行字幕生成步骤
     */
    private boolean executeSubtitleGenerate(Integer manuscriptId, Integer videoId, String videoTitle) {
        log.info("[全流程服务] 开始字幕生成: videoId={}", videoId);
        broadcastProgress(videoId, manuscriptId, videoTitle, "SUBTITLE_GENERATING", 0, Video.PROCESS_STATUS_SUBTITLE_GENERATING);
        updateVideoStatus(videoId, Video.PROCESS_STATUS_SUBTITLE_GENERATING);

        boolean success = aiSubtitleService.generateSubtitle(manuscriptId, videoId);

        if (success) {
            try {
                String subtitlePath = aiSubtitleService.getSubtitlePath(manuscriptId, videoId);
                String subtitleContent = SubtitleTextUtils.extractPlainText(subtitlePath);
                redisTemplate.opsForValue().set("subtitle:" + videoId, subtitleContent);
            } catch (Exception e) {
                log.warn("[全流程服务] 缓存字幕失败: videoId={}", videoId, e);
            }

            broadcastProgress(videoId, manuscriptId, videoTitle, "SUBTITLE_SUCCESS", 100, Video.PROCESS_STATUS_SUBTITLE_SUCCESS);
            updateVideoStatus(videoId, Video.PROCESS_STATUS_SUBTITLE_SUCCESS);
            updateVideoHasSubtitle(videoId, 1);
            log.info("[全流程服务] 字幕生成完成: videoId={}", videoId);
        } else {
            broadcastError(videoId, manuscriptId, videoTitle, "GENERATE_SUBTITLE", "字幕生成失败");
            updateVideoStatus(videoId, Video.PROCESS_STATUS_SUBTITLE_FAILED);
            log.error("[全流程服务] 字幕生成失败: videoId={}", videoId);
        }

        return success;
    }

    /**
     * 执行AI总结步骤
     */
    private boolean executeAiSummary(Integer manuscriptId, Integer videoId, String videoTitle) {
        log.info("[全流程服务] 开始AI总结: videoId={}", videoId);
        broadcastProgress(videoId, manuscriptId, videoTitle, "AI_SUMMARIZING", 0, Video.PROCESS_STATUS_AI_SUMMARIZING);
        updateVideoStatus(videoId, Video.PROCESS_STATUS_AI_SUMMARIZING);

        try {
            Video video = videoMapper.selectById(videoId);
            String subtitlePath = getSubtitlePath(manuscriptId, videoId);
            String subtitlePlainText = "";

            java.io.File subtitleFile = new java.io.File(subtitlePath);
            if (subtitleFile.exists()) {
                subtitlePlainText = SubtitleTextUtils.extractPlainText(subtitlePath);
            }

            String title = video != null ? video.getTitle() : "未知视频";

            progressSseService.pushProgress(videoId, 50, "AI分析中", "ai");
            String summary = aiSummaryService.generateSummary(subtitlePlainText, title, "");

            redisTemplate.opsForValue().set("summary:" + videoId, summary);

            broadcastProgress(videoId, manuscriptId, videoTitle, "AI_SUCCESS", 100, Video.PROCESS_STATUS_AI_SUCCESS);
            updateVideoStatus(videoId, Video.PROCESS_STATUS_AI_SUCCESS);
            updateVideoHasSummary(videoId, 1);
            log.info("[全流程服务] AI总结完成: videoId={}", videoId);

            return true;

        } catch (Exception e) {
            broadcastError(videoId, manuscriptId, videoTitle, "AI_SUMMARY", e.getMessage());
            updateVideoStatus(videoId, Video.PROCESS_STATUS_AI_FAILED);
            log.error("[全流程服务] AI总结失败: videoId={}", videoId, e);
            return false;
        }
    }

    /**
     * 广播进度
     */
    private void broadcastProgress(Integer videoId, Integer manuscriptId, String videoTitle,
                                    String stage, int progress, Integer status) {
        String stageText = getStageText(stage);
        VideoProcessWebSocketHandler.broadcastProgress(videoId, manuscriptId, videoTitle, stage, stageText, progress, status);
        progressSseService.pushProgress(videoId, progress, stageText, stage.toLowerCase());
    }

    /**
     * 广播错误
     */
    private void broadcastError(Integer videoId, Integer manuscriptId, String videoTitle,
                                 String stage, String error) {
        VideoProcessWebSocketHandler.broadcastError(videoId, manuscriptId, videoTitle, stage, error);
    }

    /**
     * 获取阶段文本
     */
    private String getStageText(String stage) {
        switch (stage) {
            case "STARTING":
                return "启动中";
            case "TRANSCODING":
                return "视频转码中";
            case "TRANSCODE_SUCCESS":
                return "转码完成";
            case "AUDIO_EXTRACTING":
                return "音频提取中";
            case "AUDIO_SUCCESS":
                return "音频提取完成";
            case "SUBTITLE_GENERATING":
                return "字幕生成中";
            case "SUBTITLE_SUCCESS":
                return "字幕生成完成";
            case "AI_SUMMARIZING":
                return "AI摘要生成中";
            case "AI_SUCCESS":
                return "AI摘要完成";
            case "COMPLETED":
                return "处理完成";
            default:
                return stage;
        }
    }

    /**
     * 更新视频状态
     */
    private void updateVideoStatus(Integer videoId, Integer status) {
        try {
            videoMapper.updateProcessStatus(videoId, status);
        } catch (Exception e) {
            log.error("[全流程服务] 更新视频状态失败: videoId={}, status={}", videoId, status, e);
        }
    }

    /**
     * 更新视频字幕标记
     */
    private void updateVideoHasSubtitle(Integer videoId, Integer hasSubtitle) {
        try {
            videoMapper.updateHasSubtitle(videoId, hasSubtitle);
        } catch (Exception e) {
            log.error("[全流程服务] 更新字幕标记失败: videoId={}", videoId, e);
        }
    }

    /**
     * 更新视频摘要标记
     */
    private void updateVideoHasSummary(Integer videoId, Integer hasSummary) {
        try {
            videoMapper.updateHasSummary(videoId, hasSummary);
        } catch (Exception e) {
            log.error("[全流程服务] 更新摘要标记失败: videoId={}", videoId, e);
        }
    }

    /**
     * 获取字幕路径
     */
    private String getSubtitlePath(Integer manuscriptId, Integer videoId) {
        return "d:/files/mybilibili/uploads/manuscripts/" + manuscriptId + "/videos/" + videoId + "/subtitles/zh-CN.srt";
    }
}
