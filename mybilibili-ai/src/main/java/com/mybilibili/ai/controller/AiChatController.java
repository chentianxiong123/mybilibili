package com.mybilibili.ai.controller;

import com.mybilibili.ai.entity.AiChatMessage;
import com.mybilibili.ai.entity.AiConversation;
import com.mybilibili.ai.service.AiChatService;
import com.mybilibili.common.vo.Result;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;

import jakarta.servlet.http.HttpServletRequest;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/ai/chat")
@Tag(name = "AI客服")
public class AiChatController {

    @Autowired
    private AiChatService aiChatService;

    @PostMapping("/send")
    @Operation(summary = "发送消息给AI客服（SSE流式返回）")
    public SseEmitter sendMessage(@RequestBody Map<String, Object> request,
                                  HttpServletRequest httpRequest) {
        Integer userId = getUserId(httpRequest);
        if (userId == null) {
            SseEmitter emitter = new SseEmitter(0L);
            try {
                emitter.send(SseEmitter.event().name("error").data("未登录"));
                emitter.complete();
            } catch (Exception ignored) {}
            return emitter;
        }

        Long conversationId = request.get("conversationId") != null
            ? Long.valueOf(request.get("conversationId").toString()) : null;
        String content = (String) request.get("content");

        if (content == null || content.trim().isEmpty()) {
            SseEmitter emitter = new SseEmitter(0L);
            try {
                emitter.send(SseEmitter.event().name("error").data("消息内容不能为空"));
                emitter.complete();
            } catch (Exception ignored) {}
            return emitter;
        }

        return aiChatService.sendMessage(userId, conversationId, content);
    }

    @GetMapping("/conversations")
    @Operation(summary = "获取用户的AI客服会话列表")
    public Result<List<AiConversation>> getConversations(HttpServletRequest httpRequest) {
        Integer userId = getUserId(httpRequest);
        if (userId == null) return Result.error("未登录");
        return Result.success(aiChatService.getConversations(userId));
    }

    @PostMapping("/conversations")
    @Operation(summary = "创建新的AI客服会话")
    public Result<AiConversation> createConversation(HttpServletRequest httpRequest) {
        Integer userId = getUserId(httpRequest);
        if (userId == null) return Result.error("未登录");
        return Result.success(aiChatService.createConversation(userId));
    }

    @DeleteMapping("/conversations/{id}")
    @Operation(summary = "删除AI客服会话")
    public Result<?> deleteConversation(@PathVariable Long id, HttpServletRequest httpRequest) {
        Integer userId = getUserId(httpRequest);
        if (userId == null) return Result.error("未登录");
        aiChatService.deleteConversation(id, userId);
        return Result.success("删除成功");
    }

    @GetMapping("/conversations/{id}/messages")
    @Operation(summary = "获取会话历史消息")
    public Result<List<AiChatMessage>> getMessages(@PathVariable Long id,
                                                    HttpServletRequest httpRequest) {
        Integer userId = getUserId(httpRequest);
        if (userId == null) return Result.error("未登录");
        return Result.success(aiChatService.getMessages(id, userId));
    }

    private Integer getUserId(HttpServletRequest request) {
        String userIdStr = request.getHeader("X-User-Id");
        if (userIdStr != null) {
            try {
                return Integer.parseInt(userIdStr);
            } catch (NumberFormatException ignored) {}
        }
        return null;
    }
}