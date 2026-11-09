package com.mybilibili.ai.service;

import java.util.List;
import java.util.Map;

public interface CustomerSessionService {
    // 获取所有需要人工处理的会话列表（按最新消息时间倒序）
    List<Map<String, Object>> listPendingSessions();
    // 获取某个会话的完整消息历史
    List<Map<String, Object>> getSessionMessages(Long sessionId);
    // 客服发送消息（人工回复）
    void sendHumanReply(Long sessionId, Long adminId, String content);
    // 将会话标记为已处理
    void markProcessed(Long sessionId, Long adminId);
    // 获取当前排队等待人工的会话数
    long countPending();
}
