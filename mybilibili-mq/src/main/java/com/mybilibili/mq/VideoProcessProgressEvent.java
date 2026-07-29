package com.mybilibili.mq;

import lombok.Data;

import java.io.Serializable;
import java.time.LocalDateTime;

@Data
public class VideoProcessProgressEvent implements Serializable {

    private static final long serialVersionUID = 1L;

    public static final String EVENT_PROGRESS = "progress";
    public static final String EVENT_ERROR = "error";
    public static final String EVENT_COMPLETE = "complete";

    private Integer videoId;
    private Integer manuscriptId;
    private String videoTitle;
    private String title;
    private String stage;
    private String stageText;
    private Integer progress;
    private Integer status;
    private String statusText;
    private String error;
    private String eventName;
    private String type;
    private boolean done;
    private String sourceService;
    private LocalDateTime occurredAt;

    public static VideoProcessProgressEvent of(Integer videoId,
                                               Integer manuscriptId,
                                               String videoTitle,
                                               String stage,
                                               String stageText,
                                               Integer progress,
                                               Integer status,
                                               String error,
                                               String eventName,
                                               String sourceService) {
        VideoProcessProgressEvent event = new VideoProcessProgressEvent();
        event.setVideoId(videoId);
        event.setManuscriptId(manuscriptId);
        event.setVideoTitle(videoTitle);
        event.setTitle(videoTitle);
        event.setStage(stage);
        event.setStageText(stageText);
        event.setProgress(progress == null ? 0 : progress);
        event.setStatus(status);
        event.setStatusText(statusText(status));
        event.setError(error);
        event.setEventName(eventName);
        event.setType(eventName);
        event.setDone(EVENT_COMPLETE.equals(eventName));
        event.setSourceService(sourceService);
        event.setOccurredAt(LocalDateTime.now());
        return event;
    }

    public static VideoProcessProgressEvent progress(Integer videoId,
                                                     Integer manuscriptId,
                                                     String videoTitle,
                                                     String stage,
                                                     String stageText,
                                                     Integer progress,
                                                     Integer status,
                                                     String sourceService) {
        return of(videoId, manuscriptId, videoTitle, stage, stageText, progress, status, null, EVENT_PROGRESS, sourceService);
    }

    public static VideoProcessProgressEvent error(Integer videoId,
                                                  Integer manuscriptId,
                                                  String videoTitle,
                                                  String stage,
                                                  String stageText,
                                                  Integer status,
                                                  String error,
                                                  String sourceService) {
        return of(videoId, manuscriptId, videoTitle, stage, stageText, 0, status, error, EVENT_ERROR, sourceService);
    }

    public static VideoProcessProgressEvent complete(Integer videoId,
                                                     Integer manuscriptId,
                                                     String videoTitle,
                                                     String stage,
                                                     String stageText,
                                                     Integer status,
                                                     String sourceService) {
        return of(videoId, manuscriptId, videoTitle, stage, stageText, 100, status, null, EVENT_COMPLETE, sourceService);
    }

    private static String statusText(Integer status) {
        if (status == null) {
            return "处理中";
        }
        switch (status) {
            case 0: return "待处理";
            case 1: return "视频转码中";
            case 10: return "转码失败";
            case 11: return "转码成功";
            case 2: return "音频提取中";
            case 20: return "音频提取失败";
            case 21: return "音频提取成功";
            case 3: return "字幕生成中";
            case 30: return "字幕生成失败";
            case 31: return "字幕生成成功";
            case 4: return "AI总结中";
            case 40: return "AI总结失败";
            case 41: return "AI总结成功";
            case 5: return "处理完成";
            default: return "未知(" + status + ")";
        }
    }
}
