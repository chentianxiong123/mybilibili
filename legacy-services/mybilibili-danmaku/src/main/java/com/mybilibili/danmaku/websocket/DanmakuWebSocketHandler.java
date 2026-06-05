package com.mybilibili.danmaku.websocket;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.danmaku.dto.SendDanmakuDTO;
import com.mybilibili.danmaku.service.DanmakuService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import org.springframework.web.socket.*;
import org.springframework.web.socket.handler.TextWebSocketHandler;
import org.springframework.web.util.UriComponentsBuilder;

import java.util.Map;

@Component
public class DanmakuWebSocketHandler extends TextWebSocketHandler {

    private static final Logger log = LoggerFactory.getLogger(DanmakuWebSocketHandler.class);
    private static final ObjectMapper objectMapper = new ObjectMapper();

    @Autowired
    private DanmakuSessionRegistry registry;

    @Autowired
    private DanmakuBroadcastService broadcastService;

    @Autowired
    private DanmakuService danmakuService;

    @Override
    public void afterConnectionEstablished(WebSocketSession session) throws Exception {
        Integer videoId = extractVideoId(session);
        if (videoId == null) {
            session.close(CloseStatus.POLICY_VIOLATION);
            return;
        }
        registry.join(videoId, session);
        log.debug("Danmaku WS joined room: videoId={}, sessionId={}", videoId, session.getId());
        session.sendMessage(new TextMessage(objectMapper.writeValueAsString(
                Map.of("type", "joined", "videoId", videoId)
        )));
    }

    @Override
    protected void handleTextMessage(WebSocketSession session, TextMessage message) throws Exception {
        String payload = message.getPayload();
        try {
            Map<?, ?> msg = objectMapper.readValue(payload, Map.class);
            String type = (String) msg.get("type");

            if ("ping".equals(type)) {
                session.sendMessage(new TextMessage(objectMapper.writeValueAsString(Map.of("type", "pong"))));
                return;
            }

            if ("danmaku".equals(type)) {
                handleDanmakuSend(session, msg);
            }
        } catch (Exception e) {
            log.debug("Invalid danmaku WS message: {}", payload);
        }
    }

    private void handleDanmakuSend(WebSocketSession session, Map<?, ?> msg) {
        Integer userId = extractUserId(session);
        Integer videoId = registry.getVideoId(session);
        if (userId == null || videoId == null) return;

        String content = (String) msg.get("content");
        Object timeObj = msg.get("time");
        String color = (String) msg.get("color");
        Number modeNum = (Number) msg.get("mode");

        if (content == null || content.isBlank()) return;

        SendDanmakuDTO dto = new SendDanmakuDTO();
        dto.setVideoId(videoId);
        dto.setContent(content);
        dto.setTime(timeObj != null ? timeObj.toString() : "0");
        dto.setColor(color != null ? color : "#ffffff");
        dto.setMode(modeNum != null ? modeNum.intValue() : 0);

        try {
            danmakuService.sendDanmaku(dto, userId);
        } catch (Exception e) {
            log.debug("Danmaku save failed: {}", e.getMessage());
            return;
        }

        double time = 0;
        try { time = Double.parseDouble(dto.getTime()); } catch (Exception ignored) {}

        broadcastService.broadcast(videoId, Map.of(
                "type", "danmaku",
                "videoId", videoId,
                "userId", userId,
                "content", content,
                "time", time,
                "color", dto.getColor(),
                "mode", dto.getMode(),
                "timestamp", System.currentTimeMillis()
        ));
    }

    @Override
    public void afterConnectionClosed(WebSocketSession session, CloseStatus status) {
        registry.leave(session);
    }

    @Override
    public void handleTransportError(WebSocketSession session, Throwable exception) {
        registry.leave(session);
    }

    private Integer extractVideoId(WebSocketSession session) {
        try {
            String query = session.getUri().getQuery();
            if (query == null) return null;
            String videoIdStr = UriComponentsBuilder.newInstance()
                    .query(query).build().getQueryParams().getFirst("videoId");
            return videoIdStr != null ? Integer.parseInt(videoIdStr) : null;
        } catch (Exception e) {
            return null;
        }
    }

    private Integer extractUserId(WebSocketSession session) {
        try {
            String query = session.getUri().getQuery();
            if (query == null) return null;
            String token = UriComponentsBuilder.newInstance()
                    .query(query).build().getQueryParams().getFirst("token");
            if (token == null || token.isBlank()) return null;
            return JwtUtils.getUserIdFromToken(token);
        } catch (Exception e) {
            return null;
        }
    }
}
