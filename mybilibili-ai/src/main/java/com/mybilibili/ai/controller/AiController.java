package com.mybilibili.ai.controller;

import com.mybilibili.ai.mapper.VideoMapper;
import com.mybilibili.ai.service.AiSummaryService;
import com.mybilibili.common.entity.Video;
import com.mybilibili.common.storage.StorageKeys;
import com.mybilibili.common.storage.StorageService;
import com.mybilibili.common.vo.Result;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.http.MediaType;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;

import java.io.IOException;
import java.io.InputStream;
import java.nio.charset.StandardCharsets;

@Slf4j
@RestController
@RequestMapping("/ai")
@Tag(name = "AI服务", description = "AI字幕生成和摘要服务")
public class AiController {

    @Autowired
    private VideoMapper videoMapper;

    @Autowired
    private AiSummaryService aiSummaryService;

    @Autowired
    private StringRedisTemplate redisTemplate;

    @Autowired
    private StorageService storageService;

    @GetMapping("/subtitle/{videoId}")
    @Operation(summary = "获取字幕内容", description = "从Redis获取生成的字幕内容")
    public Result<String> getSubtitle(@PathVariable Integer videoId) {
        String subtitle = redisTemplate.opsForValue().get("subtitle:" + videoId);
        if (subtitle == null) {
            return Result.error(404, "字幕尚未生成");
        }
        return Result.success(subtitle);
    }

    @GetMapping("/summary/{videoId}")
    @Operation(summary = "获取摘要内容", description = "从Redis获取生成的摘要内容")
    public Result<String> getSummary(@PathVariable Integer videoId) {
        String summary = redisTemplate.opsForValue().get("summary:" + videoId);
        if (summary == null) {
            return Result.error(404, "摘要尚未生成");
        }
        return Result.success(summary);
    }

    @GetMapping("/summary/stream/{videoId}")
    @Operation(summary = "流式获取视频AI摘要", description = "使用SSE流式输出视频AI摘要")
    public SseEmitter streamSummary(@PathVariable Integer videoId) {
        SseEmitter emitter = new SseEmitter(120000L);

        new Thread(() -> {
            try {
                Video video = videoMapper.selectById(videoId);
                if (video == null) {
                    emitter.send(SseEmitter.event()
                        .name("error")
                        .data("视频不存在"));
                    emitter.complete();
                    return;
                }

                if (video.getHasSummary() == null || video.getHasSummary() != 1) {
                    emitter.send(SseEmitter.event()
                        .name("error")
                        .data("该视频暂无AI摘要"));
                    emitter.complete();
                    return;
                }

                emitter.send(SseEmitter.event()
                    .name("start")
                    .data("开始生成摘要..."));

                String summaryContent = readSummaryFromStorage(videoId, video.getManuscriptId());
                if (summaryContent == null || summaryContent.isEmpty()) {
                    emitter.send(SseEmitter.event()
                        .name("error")
                        .data("摘要文件不存在或为空"));
                    emitter.complete();
                    return;
                }

                int totalLength = summaryContent.length();
                int chunkSize = 5;
                int position = 0;

                emitter.send(SseEmitter.event()
                    .name("meta")
                    .data("{\"totalLength\":" + totalLength + "}"));

                while (position < totalLength) {
                    int end = Math.min(position + chunkSize, totalLength);
                    String chunk = summaryContent.substring(position, end);

                    String encodedChunk = java.util.Base64.getEncoder().encodeToString(chunk.getBytes("UTF-8"));
                    emitter.send(SseEmitter.event()
                        .name("data")
                        .data(encodedChunk));

                    position = end;

                    int delay = 25 + (int)(Math.random() * 40);
                    Thread.sleep(delay);

                    if (position % 60 == 0 && Math.random() > 0.6) {
                        Thread.sleep(80 + (int)(Math.random() * 100));
                    }
                }

                emitter.send(SseEmitter.event()
                    .name("done")
                    .data("摘要生成完成"));

                emitter.complete();

            } catch (Exception e) {
                try {
                    emitter.send(SseEmitter.event()
                        .name("error")
                        .data("生成摘要失败: " + e.getMessage()));
                    emitter.complete();
                } catch (IOException ex) {
                    emitter.completeWithError(ex);
                }
            }
        }).start();

        return emitter;
    }

    @GetMapping("/summary/check/{videoId}")
    @Operation(summary = "检查视频是否有AI摘要", description = "快速检查视频是否已生成AI摘要")
    public Result<Boolean> checkSummary(@PathVariable Integer videoId) {
        try {
            Video video = videoMapper.selectById(videoId);
            if (video == null) {
                return Result.error("视频不存在");
            }

            boolean hasSummary = video.getHasSummary() != null && video.getHasSummary() == 1;
            return Result.success(hasSummary);

        } catch (Exception e) {
            return Result.error("检查失败: " + e.getMessage());
        }
    }

    @PostMapping("/test-api")
    @Operation(summary = "测试AI API连接", description = "测试DeepSeek API是否配置正确")
    public Result<Object> testApiConnection(@RequestBody(required = false) java.util.Map<String, String> request) {
        String testText = request != null ? request.get("text") : null;
        AiSummaryService.TestResult testResult = aiSummaryService.testApiConnection(testText);

        java.util.Map<String, Object> data = new java.util.HashMap<>();
        data.put("success", testResult.isSuccess());
        data.put("message", testResult.getMessage());
        data.put("response", testResult.getResponse());
        data.put("responseTime", testResult.getResponseTime() + "ms");

        if (testResult.isSuccess()) {
            return Result.success(data);
        } else {
            return Result.error(500, testResult.getMessage());
        }
    }

    private String readSummaryFromStorage(Integer videoId, Integer manuscriptId) {
        String key = StorageKeys.videoSummary(manuscriptId, videoId);
        try {
            String content;
            try (InputStream input = storageService.download(key)) {
                content = new String(input.readAllBytes(), StandardCharsets.UTF_8);
            }
            return extractSummaryContent(content);
        } catch (Exception e) {
            log.warn("读取摘要对象失败: {}, {}", key, e.getMessage());
            return null;
        }
    }

    private String extractSummaryContent(String fileContent) {
        if (fileContent == null || fileContent.isEmpty()) {
            return "";
        }

        int startIndex = -1;
        String[] markers = {"【视频摘要】", "### 视频摘要", "视频摘要", "### 摘要"};

        for (String marker : markers) {
            startIndex = fileContent.indexOf(marker);
            if (startIndex != -1) {
                String content = fileContent.substring(startIndex);
                return content;
            }
        }

        String[] lines = fileContent.split("\n");
        StringBuilder result = new StringBuilder();
        boolean foundEmptyLine = false;

        for (String line : lines) {
            if (line.startsWith("=") || line.startsWith("视频标题:") || line.startsWith("生成时间:")) {
                continue;
            }
            if (line.trim().isEmpty()) {
                foundEmptyLine = true;
                continue;
            }
            if (foundEmptyLine) {
                if (result.length() > 0) {
                    result.append("\n");
                }
                result.append(line);
            }
        }

        String extracted = result.toString();
        return extracted.isEmpty() ? fileContent : extracted;
    }
}
