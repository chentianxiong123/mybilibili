package com.mybilibili.mq;

import lombok.Data;

import java.io.Serializable;
import java.time.LocalDateTime;

@Data
public class VideoProcessAnalyticsEvent implements Serializable {

    private static final long serialVersionUID = 1L;

    private Integer videoId;
    private Integer manuscriptId;
    private Integer fromStatus;
    private Integer toStatus;
    private String stage;
    private Integer progress;
    private String errorMessage;
    private String operatorType;
    private Integer operatorId;
    private LocalDateTime occurredAt;

    public static VideoProcessAnalyticsEvent of(Integer videoId,
                                                Integer manuscriptId,
                                                Integer fromStatus,
                                                Integer toStatus,
                                                String stage,
                                                Integer progress,
                                                String errorMessage,
                                                String operatorType,
                                                Integer operatorId) {
        VideoProcessAnalyticsEvent event = new VideoProcessAnalyticsEvent();
        event.setVideoId(videoId);
        event.setManuscriptId(manuscriptId);
        event.setFromStatus(fromStatus);
        event.setToStatus(toStatus);
        event.setStage(stage);
        event.setProgress(progress);
        event.setErrorMessage(errorMessage);
        event.setOperatorType(operatorType);
        event.setOperatorId(operatorId);
        event.setOccurredAt(LocalDateTime.now());
        return event;
    }
}
