package com.mybilibili.danmaku.config;

import com.mybilibili.danmaku.websocket.DanmakuWebSocketHandler;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.socket.config.annotation.EnableWebSocket;
import org.springframework.web.socket.config.annotation.WebSocketConfigurer;
import org.springframework.web.socket.config.annotation.WebSocketHandlerRegistry;

@Configuration
@EnableWebSocket
public class WebSocketConfig implements WebSocketConfigurer {

    private final DanmakuWebSocketHandler danmakuHandler;

    public WebSocketConfig(DanmakuWebSocketHandler danmakuHandler) {
        this.danmakuHandler = danmakuHandler;
    }

    @Override
    public void registerWebSocketHandlers(WebSocketHandlerRegistry registry) {
        registry.addHandler(danmakuHandler, "/ws/danmaku")
                .setAllowedOrigins("*");
    }
}
