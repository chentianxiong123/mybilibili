package com.mybilibili.ai.controller;

import com.mybilibili.ai.service.AdminAiService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;

import jakarta.servlet.http.HttpServletRequest;
import java.util.Map;

@RestController
@RequestMapping("/ai/admin/assistant")
@Tag(name = "管理后台 AI 助手")
public class AdminAiController {

    @Autowired
    private AdminAiService adminAiService;

    @PostMapping("/send")
    @Operation(summary = "发送消息给管理后台 AI 助手（SSE 流式）")
    public SseEmitter sendMessage(@RequestBody Map<String, String> request,
                                  HttpServletRequest httpRequest) {
        Integer adminId = getAdminId(httpRequest);
        if (adminId == null) {
            SseEmitter emitter = new SseEmitter(0L);
            try {
                emitter.send(SseEmitter.event().name("error").data("未登录"));
                emitter.complete();
            } catch (Exception ignored) {}
            return emitter;
        }

        String content = request.get("content");
        if (content == null || content.trim().isEmpty()) {
            SseEmitter emitter = new SseEmitter(0L);
            try {
                emitter.send(SseEmitter.event().name("error").data("消息内容不能为空"));
                emitter.complete();
            } catch (Exception ignored) {}
            return emitter;
        }

        return adminAiService.sendMessage(adminId, content);
    }

    private Integer getAdminId(HttpServletRequest request) {
        String adminIdStr = request.getHeader("X-Admin-Id");
        if (adminIdStr != null) {
            try {
                return Integer.parseInt(adminIdStr);
            } catch (NumberFormatException ignored) {}
        }
        return null;
    }
}