package com.mybilibili.ai.process;

import com.mybilibili.common.entity.Video;
import com.mybilibili.mq.VideoProcessMessage;

import java.util.Optional;

public enum VideoProcessStepType {
    TRANSCODE(
            Video.PROCESS_STATUS_TRANSCODING,
            Video.PROCESS_STATUS_TRANSCODE_SUCCESS,
            Video.PROCESS_STATUS_TRANSCODE_FAILED,
            "TRANSCODING",
            "视频转码中",
            "TRANSCODE_SUCCESS",
            "转码完成",
            0,
            100,
            VideoProcessMessage.PROCESS_TYPE_TRANSCODE
    ),
    EXTRACT_AUDIO(
            Video.PROCESS_STATUS_AUDIO_EXTRACTING,
            Video.PROCESS_STATUS_AUDIO_SUCCESS,
            Video.PROCESS_STATUS_AUDIO_FAILED,
            "AUDIO_EXTRACTING",
            "音频提取中",
            "AUDIO_SUCCESS",
            "音频提取完成",
            0,
            100,
            VideoProcessMessage.PROCESS_TYPE_EXTRACT_AUDIO
    ),
    GENERATE_SUBTITLE(
            Video.PROCESS_STATUS_SUBTITLE_GENERATING,
            Video.PROCESS_STATUS_SUBTITLE_SUCCESS,
            Video.PROCESS_STATUS_SUBTITLE_FAILED,
            "SUBTITLE_GENERATING",
            "字幕生成中",
            "SUBTITLE_SUCCESS",
            "字幕生成完成",
            0,
            100,
            VideoProcessMessage.PROCESS_TYPE_GENERATE_SUBTITLE
    ),
    AI_SUMMARY(
            Video.PROCESS_STATUS_AI_SUMMARIZING,
            Video.PROCESS_STATUS_COMPLETED,
            Video.PROCESS_STATUS_AI_FAILED,
            "AI_SUMMARIZING",
            "AI总结中",
            "AI_SUCCESS",
            "AI总结完成",
            0,
            100,
            VideoProcessMessage.PROCESS_TYPE_AI_SUMMARY
    );

    private final int startedStatus;
    private final int successStatus;
    private final int failedStatus;
    private final String runningStageCode;
    private final String runningStageText;
    private final String successStageCode;
    private final String successStageText;
    private final int startProgress;
    private final int successProgress;
    private final String mqProcessType;

    VideoProcessStepType(int startedStatus,
                         int successStatus,
                         int failedStatus,
                         String runningStageCode,
                         String runningStageText,
                         String successStageCode,
                         String successStageText,
                         int startProgress,
                         int successProgress,
                         String mqProcessType) {
        this.startedStatus = startedStatus;
        this.successStatus = successStatus;
        this.failedStatus = failedStatus;
        this.runningStageCode = runningStageCode;
        this.runningStageText = runningStageText;
        this.successStageCode = successStageCode;
        this.successStageText = successStageText;
        this.startProgress = startProgress;
        this.successProgress = successProgress;
        this.mqProcessType = mqProcessType;
    }

    public int getStartedStatus() {
        return startedStatus;
    }

    public int getSuccessStatus() {
        return successStatus;
    }

    public int getFailedStatus() {
        return failedStatus;
    }

    public String getRunningStageCode() {
        return runningStageCode;
    }

    public String getRunningStageText() {
        return runningStageText;
    }

    public String getSuccessStageCode() {
        return successStageCode;
    }

    public String getSuccessStageText() {
        return successStageText;
    }

    public int getStartProgress() {
        return startProgress;
    }

    public int getSuccessProgress() {
        return successProgress;
    }

    public String toMqProcessType() {
        return mqProcessType;
    }

    public Optional<VideoProcessStepType> next() {
        switch (this) {
            case TRANSCODE:
                return Optional.of(EXTRACT_AUDIO);
            case EXTRACT_AUDIO:
                return Optional.of(GENERATE_SUBTITLE);
            case GENERATE_SUBTITLE:
                return Optional.of(AI_SUMMARY);
            default:
                return Optional.empty();
        }
    }

    public static VideoProcessStepType fromMqProcessType(String processType) {
        switch (processType) {
            case VideoProcessMessage.PROCESS_TYPE_TRANSCODE:
                return TRANSCODE;
            case VideoProcessMessage.PROCESS_TYPE_EXTRACT_AUDIO:
                return EXTRACT_AUDIO;
            case VideoProcessMessage.PROCESS_TYPE_GENERATE_SUBTITLE:
                return GENERATE_SUBTITLE;
            case VideoProcessMessage.PROCESS_TYPE_AI_SUMMARY:
                return AI_SUMMARY;
            default:
                throw new IllegalArgumentException("未知处理步骤: " + processType);
        }
    }
}
