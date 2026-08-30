package com.mybilibili.videomedia.process;

import com.mybilibili.mq.VideoProcessProgressEvent;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;
import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;

import java.io.IOException;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

@Service
public class VideoProcessProgressSseService {

    private static final Logger log = LoggerFactory.getLogger(VideoProcessProgressSseService.class);
    private static final long TIMEOUT_MILLIS = 30 * 60 * 1000L;

    private final Map<String, SseEmitter> emitters = new ConcurrentHashMap<>();
    private final Map<String, Integer> emitterVideoIds = new ConcurrentHashMap<>();
    private final Map<Integer, VideoProcessProgressEvent> latestEvents = new ConcurrentHashMap<>();

    public SseEmitter createEmitter(Integer videoId) {
        String key = videoId + "-" + System.currentTimeMillis();
        SseEmitter emitter = new SseEmitter(TIMEOUT_MILLIS);
        emitters.put(key, emitter);
        emitterVideoIds.put(key, videoId);

        emitter.onCompletion(() -> remove(key));
        emitter.onTimeout(() -> remove(key));
        emitter.onError(e -> remove(key));

        VideoProcessProgressEvent latest = latestEvents.get(videoId);
        if (latest != null) {
            try {
                emitter.send(SseEmitter.event().name("snapshot").data(latest));
            } catch (IOException e) {
                remove(key);
            }
        }
        return emitter;
    }

    public void publish(VideoProcessProgressEvent event) {
        if (event == null || event.getVideoId() == null) {
            return;
        }
        latestEvents.put(event.getVideoId(), event);
        String eventName = event.getEventName() == null ? VideoProcessProgressEvent.EVENT_PROGRESS : event.getEventName();
        publishToVideoEmitters(event, eventName);
    }

    private void publishToVideoEmitters(VideoProcessProgressEvent event, String eventName) {
        for (Map.Entry<String, SseEmitter> entry : emitters.entrySet()) {
            Integer subscribedVideoId = emitterVideoIds.get(entry.getKey());
            if (!event.getVideoId().equals(subscribedVideoId)) {
                continue;
            }
            try {
                entry.getValue().send(SseEmitter.event().name(eventName).data(event));
            } catch (Exception e) {
                log.debug("视频处理进度 SSE 推送失败，移除连接: {}", entry.getKey());
                remove(entry.getKey());
            }
        }
    }

    private void remove(String key) {
        emitters.remove(key);
        emitterVideoIds.remove(key);
    }
}
