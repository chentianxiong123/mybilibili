package com.mybilibili.ai.config;

import com.mybilibili.ai.websocket.VideoProcessWebSocketHandler;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.socket.config.annotation.EnableWebSocket;
import org.springframework.web.socket.config.annotation.WebSocketConfigurer;
import org.springframework.web.socket.config.annotation.WebSocketHandlerRegistry;

@Configuration
@EnableWebSocket
public class WebSocketConfig implements WebSocketConfigurer {
    
    private final VideoProcessWebSocketHandler videoProcessWebSocketHandler;
    
    public WebSocketConfig(VideoProcessWebSocketHandler videoProcessWebSocketHandler) {
        this.videoProcessWebSocketHandler = videoProcessWebSocketHandler;
    }
    
    @Override
    public void registerWebSocketHandlers(WebSocketHandlerRegistry registry) {
        registry.addHandler(videoProcessWebSocketHandler, "/ws/video-process")
                .setAllowedOrigins("*");
    }
}
