package com.mybilibili.ai.controller;

import com.mybilibili.ai.entity.AiChatMessage;
import com.mybilibili.ai.entity.AiSession;
import com.mybilibili.ai.mapper.AiSessionMapper;
import com.mybilibili.ai.mapper.AiChatMessageMapper;
import com.mybilibili.ai.service.CustomerServiceAiService;
import com.mybilibili.common.vo.Result;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;

import java.util.Collections;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/ai/customer")
@Tag(name = "AI客服")
public class CustomerServiceAiController {

    @Autowired
    private CustomerServiceAiService customerServiceAiService;

    @Autowired
    private AiSessionMapper aiSessionMapper;

    @Autowired
    private AiChatMessageMapper aiChatMessageMapper;

    @PostMapping("/chat")
    @Operation(summary = "AI客服对话（SSE流式返回）")
    public SseEmitter chat(@RequestBody Map<String, Object> request,
                            @RequestHeader(value = "X-User-Id", required = false) Long headerUserId) {
        Long userId = headerUserId;
        if (userId == null && request.containsKey("userId")) {
            Object userIdObj = request.get("userId");
            if (userIdObj != null) {
                if (userIdObj instanceof Number) {
                    userId = ((Number) userIdObj).longValue();
                } else {
                    try {
                        userId = Long.parseLong(userIdObj.toString());
                    } catch (NumberFormatException ignored) {}
                }
            }
        }

        if (userId == null) {
            SseEmitter emitter = new SseEmitter(0L);
            try {
                emitter.send(SseEmitter.event().name("error").data("未提供用户ID"));
                emitter.complete();
            } catch (Exception ignored) {}
            return emitter;
        }

        String content = (String) request.get("content");
        if (content == null || content.trim().isEmpty()) {
            SseEmitter emitter = new SseEmitter(0L);
            try {
                emitter.send(SseEmitter.event().name("error").data("消息内容不能为空"));
                emitter.complete();
            } catch (Exception ignored) {}
            return emitter;
        }

        return customerServiceAiService.chat(userId, content);
    }

    @GetMapping("/history/{userId}")
    @Operation(summary = "获取用户最近的AI客服会话消息")
    public Result<List<AiChatMessage>> getHistory(
            @PathVariable("userId") Long userId,
            @RequestHeader(value = "X-User-Id", required = false) Long headerUserId) {
        Long effectiveUserId = headerUserId != null ? headerUserId : userId;

        // 查该用户最新的客服会话
        var sessions = aiSessionMapper.selectByTypeAndStatus("CUSTOMER_SERVICE", 0);
        AiSession latestSession = null;
        if (sessions != null) {
            for (AiSession s : sessions) {
                if (s.getUserId().equals(effectiveUserId)) {
                    latestSession = s;
                    break;
                }
            }
        }
        if (latestSession == null) {
            return Result.success(Collections.emptyList());
        }

        List<AiChatMessage> messages = aiChatMessageMapper.selectBySessionId(latestSession.getId());
        int fromIndex = Math.max(0, messages.size() - 20);
        List<AiChatMessage> recentMessages = new java.util.ArrayList<>(messages.subList(fromIndex, messages.size()));
        return Result.success(recentMessages);
    }

    @PostMapping("/transfer")
    @Operation(summary = "转人工客服")
    public Result<Void> transfer(@RequestBody Map<String, Object> request,
                                  @RequestHeader(value = "X-User-Id", required = false) Long headerUserId) {
        Long userId = headerUserId;
        if (userId == null && request.containsKey("userId")) {
            Object userIdObj = request.get("userId");
            if (userIdObj != null) {
                if (userIdObj instanceof Number) {
                    userId = ((Number) userIdObj).longValue();
                } else {
                    try {
                        userId = Long.parseLong(userIdObj.toString());
                    } catch (NumberFormatException ignored) {}
                }
            }
        }

        Long sessionId = null;
        if (request.containsKey("sessionId")) {
            Object sessionIdObj = request.get("sessionId");
            if (sessionIdObj instanceof Number) {
                sessionId = ((Number) sessionIdObj).longValue();
            } else if (sessionIdObj != null) {
                try {
                    sessionId = Long.parseLong(sessionIdObj.toString());
                } catch (NumberFormatException ignored) {}
            }
        }

        String reason = (String) request.getOrDefault("reason", "用户主动转人工");

        try {
            customerServiceAiService.transferToHuman(sessionId, userId, reason);
            return Result.success(null);
        } catch (RuntimeException e) {
            return Result.fail(400, e.getMessage());
        }
    }
}