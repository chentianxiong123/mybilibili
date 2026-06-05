package com.mybilibili.mq;

import lombok.Data;
import java.io.Serializable;

@Data
public class VideoProcessMessage implements Serializable {

    private static final long serialVersionUID = 1L;

    private Integer manuscriptId;

    private Integer videoId;

    private Integer uploaderId;

    private String processType;

    private String processMode;

    private Integer priority;

    public static final String PROCESS_TYPE_TRANSCODE = "TRANSCODE";
    public static final String PROCESS_TYPE_EXTRACT_AUDIO = "EXTRACT_AUDIO";
    public static final String PROCESS_TYPE_GENERATE_SUBTITLE = "GENERATE_SUBTITLE";
    public static final String PROCESS_TYPE_AI_SUMMARY = "AI_SUMMARY";
    public static final String PROCESS_TYPE_ALL = "ALL";

    public static final String PROCESS_MODE_AUTO_CHAIN = "AUTO_CHAIN";
    public static final String PROCESS_MODE_MANUAL_SINGLE = "MANUAL_SINGLE";

    public static VideoProcessMessage of(Integer manuscriptId, Integer videoId, Integer uploaderId, String processType) {
        VideoProcessMessage message = new VideoProcessMessage();
        message.setManuscriptId(manuscriptId);
        message.setVideoId(videoId);
        message.setUploaderId(uploaderId);
        message.setProcessType(processType);
        message.setProcessMode(PROCESS_MODE_AUTO_CHAIN);
        message.setPriority(5);
        return message;
    }

    public static VideoProcessMessage manualSingle(Integer manuscriptId, Integer videoId, Integer uploaderId, String processType) {
        VideoProcessMessage message = of(manuscriptId, videoId, uploaderId, processType);
        message.setProcessMode(PROCESS_MODE_MANUAL_SINGLE);
        return message;
    }
}
