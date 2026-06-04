package com.mybilibili.ai.pipeline;

import lombok.Data;
import java.io.Serializable;
import java.time.LocalDateTime;
import java.util.concurrent.atomic.AtomicInteger;

/**
 * 全流程处理任务
 */
@Data
public class PipelineTask implements Serializable {

    private static final long serialVersionUID = 1L;

    private Integer manuscriptId;
    private Integer videoId;
    private Integer uploaderId;
    private String videoTitle;
    private Long persistentId;

    /**
     * 当前步骤索引
     */
    private AtomicInteger currentStepIndex = new AtomicInteger(0);

    /**
     * 任务状态
     */
    private TaskStatus status = TaskStatus.PENDING;

    /**
     * 创建时间
     */
    private LocalDateTime createTime;

    /**
     * 开始时间
     */
    private LocalDateTime startTime;

    /**
     * 结束时间
     */
    private LocalDateTime endTime;

    /**
     * 错误信息
     */
    private String errorMessage;

    /**
     * 失败步骤
     */
    private String failedStep;

    /**
     * 任务状态枚举
     */
    public enum TaskStatus {
        PENDING("待处理"),
        RUNNING("处理中"),
        COMPLETED("已完成"),
        FAILED("失败"),
        CANCELLED("已取消");

        private final String description;

        TaskStatus(String description) {
            this.description = description;
        }

        public String getDescription() {
            return description;
        }
    }

    /**
     * 处理步骤枚举
     */
    public enum ProcessStep {
        TRANSCODE(0, "转码", "TRANSCODE"),
        AUDIO_EXTRACT(1, "音频提取", "EXTRACT_AUDIO"),
        SUBTITLE_GENERATE(2, "字幕生成", "GENERATE_SUBTITLE"),
        AI_SUMMARY(3, "AI总结", "AI_SUMMARY");

        private final int index;
        private final String description;
        private final String code;

        ProcessStep(int index, String description, String code) {
            this.index = index;
            this.description = description;
            this.code = code;
        }

        public int getIndex() {
            return index;
        }

        public String getDescription() {
            return description;
        }

        public String getCode() {
            return code;
        }

        public static ProcessStep fromIndex(int index) {
            for (ProcessStep step : values()) {
                if (step.index == index) {
                    return step;
                }
            }
            return null;
        }

        public static ProcessStep fromCode(String code) {
            for (ProcessStep step : values()) {
                if (step.code.equals(code)) {
                    return step;
                }
            }
            return null;
        }
    }

    /**
     * 创建任务
     */
    public static PipelineTask create(Integer manuscriptId, Integer videoId, Integer uploaderId, String videoTitle) {
        PipelineTask task = new PipelineTask();
        task.setManuscriptId(manuscriptId);
        task.setVideoId(videoId);
        task.setUploaderId(uploaderId);
        task.setVideoTitle(videoTitle);
        task.setCreateTime(LocalDateTime.now());
        task.setStatus(TaskStatus.PENDING);
        return task;
    }

    /**
     * 获取当前步骤
     */
    public ProcessStep getCurrentStep() {
        return ProcessStep.fromIndex(currentStepIndex.get());
    }

    /**
     * 前进到下一步
     */
    public boolean advanceToNextStep() {
        int nextIndex = currentStepIndex.incrementAndGet();
        return nextIndex < ProcessStep.values().length;
    }

    /**
     * 是否已完成所有步骤
     */
    public boolean isAllStepsCompleted() {
        return currentStepIndex.get() >= ProcessStep.values().length;
    }

    /**
     * 标记任务开始
     */
    public void markStarted() {
        this.status = TaskStatus.RUNNING;
        this.startTime = LocalDateTime.now();
    }

    /**
     * 标记任务完成
     */
    public void markCompleted() {
        this.status = TaskStatus.COMPLETED;
        this.endTime = LocalDateTime.now();
    }

    /**
     * 标记任务失败
     */
    public void markFailed(String step, String error) {
        this.status = TaskStatus.FAILED;
        this.failedStep = step;
        this.errorMessage = error;
        this.endTime = LocalDateTime.now();
    }

    /**
     * 标记任务取消
     */
    public void markCancelled() {
        this.status = TaskStatus.CANCELLED;
        this.endTime = LocalDateTime.now();
    }

    /**
     * 获取任务唯一标识
     */
    public String getTaskKey() {
        return manuscriptId + "-" + videoId;
    }

    @Override
    public String toString() {
        return String.format("PipelineTask{videoId=%d, manuscriptId=%d, status=%s, step=%s}",
                videoId, manuscriptId, status, getCurrentStep());
    }
}
