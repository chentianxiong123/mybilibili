package com.mybilibili.ai.websocket;

import com.fasterxml.jackson.databind.ObjectMapper;
import org.springframework.stereotype.Component;
import org.springframework.web.socket.CloseStatus;
import org.springframework.web.socket.TextMessage;
import org.springframework.web.socket.WebSocketSession;
import org.springframework.web.socket.handler.TextWebSocketHandler;

import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.CopyOnWriteArraySet;

@Component
public class VideoProcessWebSocketHandler extends TextWebSocketHandler {
    
    private static final CopyOnWriteArraySet<WebSocketSession> sessions = new CopyOnWriteArraySet<>();
    private static final ObjectMapper objectMapper = new ObjectMapper();
    
    @Override
    public void afterConnectionEstablished(WebSocketSession session) {
        sessions.add(session);
        System.out.println("[WebSocket] 新连接，当前连接数: " + sessions.size());
    }
    
    @Override
    public void afterConnectionClosed(WebSocketSession session, CloseStatus status) {
        sessions.remove(session);
        System.out.println("[WebSocket] 连接关闭，当前连接数: " + sessions.size());
    }
    
    @Override
    protected void handleTextMessage(WebSocketSession session, TextMessage message) throws Exception {
        System.out.println("[WebSocket] 收到消息: " + message.getPayload());
    }
    
    public static void broadcast(Map<String, Object> data) {
        try {
            String json = objectMapper.writeValueAsString(data);
            TextMessage message = new TextMessage(json);
            for (WebSocketSession session : sessions) {
                if (session.isOpen()) {
                    session.sendMessage(message);
                }
            }
        } catch (Exception e) {
            System.err.println("[WebSocket] 广播失败: " + e.getMessage());
        }
    }
    
    public static void broadcastProgress(Integer videoId, Integer manuscriptId, String videoTitle,
                                          String stage, String stageText, int progress, Integer status) {
        Map<String, Object> data = new HashMap<>();
        data.put("type", "progress");
        data.put("videoId", videoId);
        data.put("manuscriptId", manuscriptId);
        data.put("videoTitle", videoTitle);
        data.put("stage", stage);
        data.put("stageText", stageText);
        data.put("progress", progress);
        data.put("status", status);
        broadcast(data);
    }
    
    public static void broadcastComplete(Integer videoId, Integer manuscriptId, String videoTitle) {
        Map<String, Object> data = new HashMap<>();
        data.put("type", "complete");
        data.put("videoId", videoId);
        data.put("manuscriptId", manuscriptId);
        data.put("videoTitle", videoTitle);
        broadcast(data);
    }
    
    public static void broadcastError(Integer videoId, Integer manuscriptId, String videoTitle,
                                       String stage, String error) {
        Map<String, Object> data = new HashMap<>();
        data.put("type", "error");
        data.put("videoId", videoId);
        data.put("manuscriptId", manuscriptId);
        data.put("videoTitle", videoTitle);
        data.put("stage", stage);
        data.put("error", error);
        broadcast(data);
    }
}
