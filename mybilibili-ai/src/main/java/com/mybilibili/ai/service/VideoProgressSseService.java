package com.mybilibili.ai.service;

import com.mybilibili.common.entity.Video;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;
import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;

import java.io.IOException;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

@Service
public class VideoProgressSseService {

    private static final Logger log = LoggerFactory.getLogger(VideoProgressSseService.class);

    private final Map<String, SseEmitter> emitters = new ConcurrentHashMap<>();
    private final Map<Integer, String> videoTitles = new ConcurrentHashMap<>();

    public SseEmitter createEmitter(Integer videoId) {
        String key = videoId + "-" + System.currentTimeMillis();
        log.info("[SSE创建] videoId={}, key={}", videoId, key);
        SseEmitter emitter = new SseEmitter(30 * 60 * 1000L);
        emitters.put(key, emitter);
        emitter.onCompletion(() -> {
            log.info("[SSE完成] key={}", key);
            emitters.remove(key);
        });
        emitter.onTimeout(() -> {
            log.info("[SSE超时] key={}", key);
            emitters.remove(key);
        });
        emitter.onError(e -> {
            log.info("[SSE错误] key={}, error={}", key, e.getMessage());
            emitters.remove(key);
        });
        return emitter;
    }

    public SseEmitter createAdminEmitter(Map<String, Object> snapshot) {
        String key = "admin-" + System.currentTimeMillis();
        SseEmitter emitter = new SseEmitter(30 * 60 * 1000L);
        emitters.put(key, emitter);
        emitter.onCompletion(() -> emitters.remove(key));
        emitter.onTimeout(() -> emitters.remove(key));
        emitter.onError(e -> emitters.remove(key));
        try {
            emitter.send(SseEmitter.event().name("snapshot").data(snapshot));
        } catch (IOException e) {
            emitters.remove(key);
        }
        return emitter;
    }

    public void setVideoTitle(Integer videoId, String title) {
        videoTitles.put(videoId, title);
    }

    public void pushProgress(Integer videoId, Integer progress, String status, String stage) {
        String title = videoTitles.getOrDefault(videoId, "");
        Map<String, Object> data = buildEvent(videoId, null, title, stage, status, progress, null, null);
        broadcast("progress", data);
    }

    public void pushProgress(Integer videoId, Integer progress, String status, String stage, String title) {
        Map<String, Object> data = buildEvent(videoId, null, title, stage, status, progress, null, null);
        broadcast("progress", data);
    }

    public void pushProcessEvent(Video video, String eventName, String stageText) {
        if (video == null) {
            return;
        }
        Map<String, Object> data = buildEvent(
                video.getId(),
                video.getManuscriptId(),
                video.getTitle(),
                video.getProcessStage(),
                stageText,
                video.getProcessProgress(),
                video.getProcessStatus(),
                video.getProcessError()
        );
        broadcast(eventName, data);
    }

    public void complete(Integer videoId) {
        Map<String, Object> data = new HashMap<>();
        data.put("videoId", videoId);
        data.put("done", true);
        broadcast("complete", data);
    }

    private Map<String, Object> buildEvent(Integer videoId, Integer manuscriptId, String title, String stage,
                                           String stageText, Integer progress, Integer status, String error) {
        Map<String, Object> data = new HashMap<>();
        data.put("type", eventType(status, error));
        data.put("videoId", videoId);
        data.put("manuscriptId", manuscriptId);
        data.put("videoTitle", title);
        data.put("title", title);
        data.put("stage", stage);
        data.put("stageText", stageText);
        data.put("progress", progress == null ? 0 : progress);
        data.put("status", status);
        data.put("statusText", statusText(status));
        data.put("error", error);
        return data;
    }

    private String eventType(Integer status, String error) {
        if (error != null && !error.isEmpty()) {
            return "error";
        }
        if (status != null && status == Video.PROCESS_STATUS_COMPLETED) {
            return "complete";
        }
        return "progress";
    }

    private String statusText(Integer status) {
        if (status == null) {
            return "处理中";
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
            default: return "未知(" + status + ")";
        }
    }

    private void broadcast(String eventName, Object data) {
        log.info("[SSE广播] event={}, emitterCount={}", eventName, emitters.size());
        for (Map.Entry<String, SseEmitter> entry : emitters.entrySet()) {
            try {
                entry.getValue().send(SseEmitter.event().name(eventName).data(data));
            } catch (Exception e) {
                log.error("[SSE广播] 失败 emitter={}, error={}", entry.getKey(), e.getMessage());
                emitters.remove(entry.getKey());
            }
        }
    }
}
