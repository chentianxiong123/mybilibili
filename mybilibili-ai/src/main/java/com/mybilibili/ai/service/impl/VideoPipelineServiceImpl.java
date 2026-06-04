package com.mybilibili.ai.service.impl;

import com.mybilibili.ai.entity.VideoPipelineTaskEntity;
import com.mybilibili.ai.mapper.VideoPipelineTaskMapper;
import com.mybilibili.ai.mapper.VideoMapper;
import com.mybilibili.ai.pipeline.PipelineTask;
import com.mybilibili.ai.repository.SubtitleRepository;
import com.mybilibili.ai.service.*;
import com.mybilibili.ai.utils.SubtitleTextUtils;
import com.mybilibili.ai.websocket.VideoProcessWebSocketHandler;
import com.mybilibili.common.entity.Subtitle;
import com.mybilibili.common.entity.Video;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;

import jakarta.annotation.PostConstruct;
import jakarta.annotation.PreDestroy;
import java.net.InetAddress;
import java.nio.file.Path;
import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.ScheduledFuture;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.stream.Collectors;

/**
 * 视频全流程处理服务实现
 * 实现自动按顺序执行视频的所有处理步骤（转码→音频→字幕→AI）
 */
@Service
public class VideoPipelineServiceImpl implements VideoPipelineService {

    private static final Logger log = LoggerFactory.getLogger(VideoPipelineServiceImpl.class);

    private static final int TASK_LEASE_SECONDS = 3600;

    private final String workerId = buildWorkerId();

    private final ExecutorService pipelineExecutor = Executors.newSingleThreadExecutor(r -> {
        Thread thread = new Thread(r, "video-pipeline-db-worker");
        thread.setDaemon(true);
        return thread;
    });

    private final ScheduledExecutorService leaseScheduler = Executors.newSingleThreadScheduledExecutor(r -> {
        Thread thread = new Thread(r, "video-pipeline-lease-renewer");
        thread.setDaemon(true);
        return thread;
    });

    private volatile boolean workerRunning = true;

    private volatile PipelineTask currentTask;
    private volatile ScheduledFuture<?> leaseRenewalFuture;

    @Autowired
    private VideoPipelineTaskMapper pipelineTaskMapper;
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

    @Autowired
    private SubtitleRepository subtitleRepository;

    @Autowired
    private VideoProcessingStorageService processingStorageService;

    /**
     * 初始化：启动基于数据库租约的单线程任务处理器
     */
    @PostConstruct
    public void init() {
        // 实验性架构：暂时关掉轮询 worker，任务由 API 手动触发；如需恢复调度取消下一行注释。
        // pipelineExecutor.submit(this::processTaskLoop);
        log.info("[全流程服务] 初始化完成，轮询 worker 已禁用（实验模式），workerId={}", workerId);
    }

    @Override
    public boolean submitPipelineTask(Integer manuscriptId, Integer videoId, Integer uploaderId) {
        // 获取视频信息
        Video video = videoMapper.selectById(videoId);
        String videoTitle = video != null ? video.getTitle() : "未知视频";

        // 创建任务
        PipelineTask task = PipelineTask.create(manuscriptId, videoId, uploaderId, videoTitle);
        String taskKey = task.getTaskKey();
        if (pipelineTaskMapper.countActiveByTaskKey(taskKey) > 0) {
            log.warn("[全流程服务] 视频处理任务已存在: {}", taskKey);
            return false;
        }
        VideoPipelineTaskEntity entity = toEntity(task);
        pipelineTaskMapper.insert(entity);
        task.setPersistentId(entity.getId());

        // 设置视频标题到进度服务
        progressSseService.setVideoTitle(videoId, videoTitle);

        log.info("[全流程服务] 任务已提交: manuscriptId={}, videoId={}, title={}",
                manuscriptId, videoId, videoTitle);
        updateVideoStatus(videoId, Video.PROCESS_STATUS_TRANSCODING);
        return true;
    }

    @Override
    public boolean cancelTask(Integer manuscriptId, Integer videoId) {
        String taskKey = manuscriptId + "-" + videoId;
        VideoPipelineTaskEntity task = pipelineTaskMapper.selectLatestByTaskKey(taskKey);
        if (task == null || !"PENDING".equals(task.getStatus())) {
            return false;
        }
        return pipelineTaskMapper.markCancelled(task.getId()) > 0;
    }

    @Override
    public PipelineTask getTaskStatus(Integer manuscriptId, Integer videoId) {
        String taskKey = manuscriptId + "-" + videoId;
        return toTask(pipelineTaskMapper.selectLatestByTaskKey(taskKey));
    }

    @Override
    public Map<String, Object> getStatistics() {
        Map<String, Object> stats = new HashMap<>();
        stats.put("workerId", workerId);
        stats.put("workerRunning", workerRunning);
        stats.put("currentTask", currentTask != null ? currentTask.toString() : null);
        stats.put("dbPending", pipelineTaskMapper.countByStatus(PipelineTask.TaskStatus.PENDING.name()));
        stats.put("dbRunning", pipelineTaskMapper.countByStatus(PipelineTask.TaskStatus.RUNNING.name()));
        stats.put("dbCompleted", pipelineTaskMapper.countByStatus(PipelineTask.TaskStatus.COMPLETED.name()));
        stats.put("dbFailed", pipelineTaskMapper.countByStatus(PipelineTask.TaskStatus.FAILED.name()));
        stats.put("dbCancelled", pipelineTaskMapper.countByStatus(PipelineTask.TaskStatus.CANCELLED.name()));
        return stats;
    }

    @Override
    public List<PipelineTask> getQueuedTasks() {
        return toTasks(pipelineTaskMapper.selectByStatus(PipelineTask.TaskStatus.PENDING.name(), 100));
    }

    @Override
    public List<PipelineTask> getProcessingTasks() {
        return toTasks(pipelineTaskMapper.selectByStatus(PipelineTask.TaskStatus.RUNNING.name(), 100));
    }

    @Override
    public List<PipelineTask> getCompletedTasks() {
        return toTasks(pipelineTaskMapper.selectByStatus(PipelineTask.TaskStatus.COMPLETED.name(), 100));
    }

    @Override
    public PipelineTask getCurrentTask() {
        if (currentTask != null) {
            return currentTask;
        }
        return toTask(pipelineTaskMapper.selectClaimedTask(workerId));
    }

    @Override
    public void clearQueue() {
        pipelineTaskMapper.cancelPendingTasks();
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

    private void processTaskLoop() {
        while (workerRunning) {
            try {
                int interrupted = pipelineTaskMapper.markExpiredRunningTasksInterrupted();
                if (interrupted > 0) {
                    log.warn("[全流程服务] 已将 {} 个租约过期任务标记为失败", interrupted);
                }
                int claimed = pipelineTaskMapper.claimNextPendingTask(workerId, TASK_LEASE_SECONDS);
                if (claimed <= 0) {
                    sleepQuietly(1000);
                    continue;
                }
                PipelineTask task = toTask(pipelineTaskMapper.selectClaimedTask(workerId));
                if (task == null) {
                    log.warn("[全流程服务] 已认领任务但未查询到任务实体，workerId={}", workerId);
                    continue;
                }
                currentTask = task;
                progressSseService.setVideoTitle(task.getVideoId(), task.getVideoTitle());
                processTask(task);
            } catch (Exception e) {
                log.error("[全流程服务] 数据库任务调度循环异常", e);
                sleepQuietly(1000);
            } finally {
                currentTask = null;
            }
        }
        log.info("[全流程服务] 数据库任务调度循环已停止，workerId={}", workerId);
    }

    @PreDestroy
    public void shutdown() {
        workerRunning = false;
        stopLeaseRenewal();
        pipelineExecutor.shutdown();
        leaseScheduler.shutdown();
        try {
            if (!pipelineExecutor.awaitTermination(30, TimeUnit.SECONDS)) {
                pipelineExecutor.shutdownNow();
            }
            if (!leaseScheduler.awaitTermination(30, TimeUnit.SECONDS)) {
                leaseScheduler.shutdownNow();
            }
        } catch (InterruptedException e) {
            pipelineExecutor.shutdownNow();
            leaseScheduler.shutdownNow();
            Thread.currentThread().interrupt();
        }
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
            markTaskRunning(task);
            startLeaseRenewal(task);
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
                    markTaskFailed(task);
                    broadcastError(videoId, manuscriptId, videoTitle, currentStep.getCode(), currentStep.getDescription() + "失败");
                    return;
                }

                // 步骤成功，前进到下一步
                boolean hasMoreSteps = task.advanceToNextStep();
                updateTaskStep(task);

                if (!hasMoreSteps) {
                    // 所有步骤完成
                    task.markCompleted();
                    markTaskCompleted(task);
                    log.info("[全流程服务] 任务完成: {}", task.getTaskKey());
                    broadcastProgress(videoId, manuscriptId, videoTitle, "COMPLETED", 100, Video.PROCESS_STATUS_COMPLETED);
                    VideoProcessWebSocketHandler.broadcastComplete(videoId, manuscriptId, videoTitle);
                    progressSseService.complete(videoId);
                }
            }

        } catch (Exception e) {
            log.error("[全流程服务] 任务处理异常: {}", task.getTaskKey(), e);
            task.markFailed("UNKNOWN", e.getMessage());
            markTaskFailed(task);
            broadcastError(videoId, manuscriptId, videoTitle, "UNKNOWN", e.getMessage());
        } finally {
            stopLeaseRenewal();
            processingStorageService.cleanupWorkDir(manuscriptId, videoId);
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
            String subtitlePlainText = buildSubtitleText(videoId);
            if (subtitlePlainText.isBlank()) {
                throw new IllegalStateException("字幕内容为空，不能生成AI摘要");
            }

            String title = video != null ? video.getTitle() : "未知视频";

            progressSseService.pushProgress(videoId, 50, "AI分析中", "ai");
            String summary = aiSummaryService.generateSummary(subtitlePlainText, title, "");
            if (summary == null || summary.isBlank()) {
                throw new IllegalStateException("AI摘要生成结果为空");
            }

            Path summaryPath = processingStorageService.getSummaryPath(manuscriptId, videoId);
            if (!aiSummaryService.saveSummaryToFile(summary, summaryPath.toString(), title)) {
                throw new IllegalStateException("AI摘要保存失败");
            }
            processingStorageService.uploadSummary(manuscriptId, videoId, summaryPath);

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

    private String buildSubtitleText(Integer videoId) {
        List<Subtitle> subtitles = subtitleRepository.findByVideoId(videoId);
        if (subtitles == null || subtitles.isEmpty()) {
            return "";
        }
        StringBuilder text = new StringBuilder();
        for (Subtitle subtitle : subtitles) {
            if (subtitle.getContent() == null) {
                continue;
            }
            for (Subtitle.SubtitleItem item : subtitle.getContent()) {
                if (item.getText() != null && !item.getText().isBlank()) {
                    text.append(item.getText()).append('\n');
                }
            }
        }
        return text.toString();
    }

    private VideoPipelineTaskEntity toEntity(PipelineTask task) {
        VideoPipelineTaskEntity entity = new VideoPipelineTaskEntity();
        entity.setTaskKey(task.getTaskKey());
        entity.setManuscriptId(task.getManuscriptId());
        entity.setVideoId(task.getVideoId());
        entity.setUploaderId(task.getUploaderId());
        entity.setVideoTitle(task.getVideoTitle());
        entity.setCurrentStepIndex(task.getCurrentStepIndex() != null ? task.getCurrentStepIndex().get() : 0);
        entity.setStatus(task.getStatus().name());
        entity.setWorkerId(null);
        entity.setLockedUntil(null);
        entity.setCreateTime(task.getCreateTime() != null ? task.getCreateTime() : LocalDateTime.now());
        entity.setStartTime(task.getStartTime());
        entity.setEndTime(task.getEndTime());
        entity.setFailedStep(task.getFailedStep());
        entity.setErrorMessage(task.getErrorMessage());
        return entity;
    }

    private PipelineTask toTask(VideoPipelineTaskEntity entity) {
        if (entity == null) {
            return null;
        }
        PipelineTask task = new PipelineTask();
        task.setPersistentId(entity.getId());
        task.setManuscriptId(entity.getManuscriptId());
        task.setVideoId(entity.getVideoId());
        task.setUploaderId(entity.getUploaderId());
        task.setVideoTitle(entity.getVideoTitle());
        task.setCurrentStepIndex(new AtomicInteger(entity.getCurrentStepIndex() != null ? entity.getCurrentStepIndex() : 0));
        task.setCreateTime(entity.getCreateTime());
        task.setStartTime(entity.getStartTime());
        task.setEndTime(entity.getEndTime());
        task.setFailedStep(entity.getFailedStep());
        task.setErrorMessage(entity.getErrorMessage());
        task.setStatus(PipelineTask.TaskStatus.valueOf(entity.getStatus()));
        return task;
    }

    private void startLeaseRenewal(PipelineTask task) {
        stopLeaseRenewal();
        if (task.getPersistentId() == null) {
            return;
        }
        leaseRenewalFuture = leaseScheduler.scheduleAtFixedRate(
                () -> pipelineTaskMapper.renewLease(task.getPersistentId(), workerId, TASK_LEASE_SECONDS),
                30,
                30,
                TimeUnit.SECONDS
        );
    }

    private void stopLeaseRenewal() {
        ScheduledFuture<?> future = leaseRenewalFuture;
        if (future != null) {
            future.cancel(false);
            leaseRenewalFuture = null;
        }
    }

    private List<PipelineTask> toTasks(List<VideoPipelineTaskEntity> entities) {
        if (entities == null || entities.isEmpty()) {
            return new ArrayList<>();
        }
        return entities.stream().map(this::toTask).collect(Collectors.toList());
    }

    private void markTaskRunning(PipelineTask task) {
        if (task.getPersistentId() != null) {
            pipelineTaskMapper.markRunning(task.getPersistentId(), task.getCurrentStepIndex().get(), workerId, TASK_LEASE_SECONDS);
        }
    }

    private void updateTaskStep(PipelineTask task) {
        if (task.getPersistentId() != null) {
            pipelineTaskMapper.updateCurrentStep(task.getPersistentId(), task.getCurrentStepIndex().get());
            pipelineTaskMapper.renewLease(task.getPersistentId(), workerId, TASK_LEASE_SECONDS);
        }
    }

    private void markTaskCompleted(PipelineTask task) {
        if (task.getPersistentId() != null) {
            pipelineTaskMapper.markCompleted(task.getPersistentId(), task.getCurrentStepIndex().get());
        }
    }

    private void markTaskFailed(PipelineTask task) {
        if (task.getPersistentId() != null) {
            pipelineTaskMapper.markFailed(task.getPersistentId(), task.getFailedStep(), task.getErrorMessage());
        }
    }

    private static String buildWorkerId() {
        String host;
        try {
            host = InetAddress.getLocalHost().getHostName();
        } catch (Exception e) {
            host = "unknown-host";
        }
        return host + "-" + UUID.randomUUID();
    }

    private void sleepQuietly(long millis) {
        try {
            Thread.sleep(millis);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            workerRunning = false;
        }
    }
}
