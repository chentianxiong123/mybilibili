package com.mybilibili.videomedia.process;

import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;

@RestController
@RequestMapping("/api/video/process")
@Tag(name = "视频处理进度", description = "统一视频处理进度实时推送")
public class VideoProcessProgressController {

    private final VideoProcessProgressSseService sseService;

    public VideoProcessProgressController(VideoProcessProgressSseService sseService) {
        this.sseService = sseService;
    }

    @GetMapping("/sse/{videoId}")
    @Operation(summary = "获取视频处理进度 SSE 流")
    public SseEmitter subscribe(@PathVariable Integer videoId) {
        return sseService.createEmitter(videoId);
    }
}
