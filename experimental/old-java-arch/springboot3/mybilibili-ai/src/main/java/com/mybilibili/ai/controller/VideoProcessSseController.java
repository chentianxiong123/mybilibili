package com.mybilibili.ai.controller;

import com.mybilibili.ai.service.VideoProgressSseService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;

@RestController
@RequestMapping("/api/video/process")
@Tag(name = "视频处理SSE", description = "视频处理进度实时推送")
public class VideoProcessSseController {

    @Autowired
    private VideoProgressSseService sseService;

    @GetMapping("/sse/{videoId}")
    @Operation(summary = "获取视频处理进度SSE流")
    public SseEmitter subscribe(@PathVariable Integer videoId) {
        return sseService.createEmitter(videoId);
    }
}
