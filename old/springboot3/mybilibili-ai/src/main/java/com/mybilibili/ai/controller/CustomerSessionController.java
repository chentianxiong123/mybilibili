package com.mybilibili.ai.controller;

import com.mybilibili.ai.service.CustomerSessionService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/ai/admin/customer")
public class CustomerSessionController {

    @Autowired
    private CustomerSessionService customerSessionService;

    /**
     * 获取待处理会话列表
     */
    @GetMapping("/sessions")
    public Map<String, Object> listPendingSessions() {
        List<Map<String, Object>> sessions = customerSessionService.listPendingSessions();
        return Map.of(
            "code", 200,
            "data", sessions
        );
    }

    /**
     * 获取某个会话的消息历史
     */
    @GetMapping("/sessions/{sessionId}/messages")
    public Map<String, Object> getSessionMessages(@PathVariable Long sessionId) {
        List<Map<String, Object>> messages = customerSessionService.getSessionMessages(sessionId);
        return Map.of(
            "code", 200,
            "data", messages
        );
    }

    /**
     * 客服发送消息（人工回复）
     */
    @PostMapping("/sessions/{sessionId}/reply")
    public Map<String, Object> sendReply(@PathVariable Long sessionId, @RequestBody Map<String, Object> body) {
        Long adminId = Long.valueOf(body.get("adminId").toString());
        String content = body.get("content").toString();
        customerSessionService.sendHumanReply(sessionId, adminId, content);
        return Map.of(
            "code", 200,
            "message", "发送成功"
        );
    }

    /**
     * 标记会话为已处理
     */
    @PostMapping("/sessions/{sessionId}/resolve")
    public Map<String, Object> resolveSession(@PathVariable Long sessionId) {
        customerSessionService.markProcessed(sessionId);
        return Map.of(
            "code", 200,
            "message", "会话已标记为已处理"
        );
    }

    /**
     * 获取等待人工处理的会话数量
     */
    @GetMapping("/sessions/pending/count")
    public Map<String, Object> getPendingCount() {
        long count = customerSessionService.countPending();
        return Map.of(
            "code", 200,
            "data", count
        );
    }
}
