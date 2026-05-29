package com.mybilibili.ai.service.impl;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.mybilibili.ai.config.DynamicChatClient;
import com.mybilibili.ai.service.AdminAiService;
import com.mybilibili.ai.tool.AdminToolService;
import com.mybilibili.ai.tool.StatsData;
import com.mybilibili.ai.util.AiUsageLogger;
import org.springframework.ai.chat.client.ChatClient;
import org.springframework.ai.tool.ToolCallback;
import org.springframework.ai.support.ToolCallbacks;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;
import reactor.core.Disposable;
import reactor.core.publisher.Flux;

import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.atomic.AtomicReference;

@Service
public class AdminAiServiceImpl implements AdminAiService {

    @Autowired
    private DynamicChatClient dynamicChatClient;

    @Autowired
    private AdminToolService adminToolService;

    @Autowired
    private AiUsageLogger aiUsageLogger;

    private final ObjectMapper objectMapper = new ObjectMapper();

    private static final String SYSTEM_PROMPT =
        "你是哔哩哔哩管理后台的 AI 助手，名叫\"管理助手\"。你的职责是帮助管理员完成以下工作：1. 查询平台统计数据，并以表格、图表或文字摘要形式呈现 2. 回答关于平台运营状况的问题。你可以使用以下工具获取数据：getOverviewStats: 获取平台概览统计；getUserGrowth(days): 获取用户增长趋势；getManuscriptStats: 获取稿件状态分布统计；getAiUsageOverview: 获取 AI 功能用量概览；getHotVideos(limit): 获取热门视频列表。当用户请求数据时，调用对应工具并将结果组织成清晰的中文回答。工具返回的 StatsData 中包含 chartType 字段。请在回答末尾用 JSON 格式返回渲染建议：{\"render\": {\"type\": \"chartType\", \"title\": \"标题\", \"data\": {...}}}";

    @Override
    public SseEmitter sendMessage(Integer adminId, String content) {
        SseEmitter emitter = new SseEmitter(120000L);
        long startTime = System.currentTimeMillis();

        // 获取绑定 ADMIN feature 的 ChatClient，无则用第一个活跃渠道
        ChatClient client = dynamicChatClient.getClient("ADMIN");
        if (client == null) {
            client = dynamicChatClient.getFirstActiveClient();
        }

        if (client == null) {
            try {
                emitter.send(SseEmitter.event().name("error").data("未配置可用的 AI 渠道，请先在 AI 渠道管理中配置并绑定 ADMIN 功能"));
                emitter.complete();
            } catch (Exception ignored) {}
            return emitter;
        }

        // 注册 ToolCallback（@Tool 注解的方法会被自动扫描）
        List<ToolCallback> toolCallbacks = List.of(
            ToolCallbacks.from(adminToolService)
        );

        StringBuilder fullResp = new StringBuilder();
        AtomicReference<Disposable> subscriptionRef = new AtomicReference<>();

        Flux<String> flux = client.prompt()
                .system(SYSTEM_PROMPT)
                .user(content)
                .tools(toolCallbacks)
                .stream()
                .content();

        Disposable subscription = flux.subscribe(
            chunk -> {
                try {
                    fullResp.append(chunk);
                    emitter.send(SseEmitter.event().name("data").data(chunk));
                } catch (Exception ignored) {}
            },
            error -> {
                aiUsageLogger.log("ADMIN", null, null, null,
                    System.currentTimeMillis() - startTime, false, error.getMessage());
                try {
                    emitter.send(SseEmitter.event().name("error")
                        .data("请求失败: " + error.getMessage()));
                    emitter.complete();
                } catch (Exception ignored) {}
            },
            () -> {
                try {
                    String reply = fullResp.toString();
                    // 尝试从回复中提取 render JSON
                    String renderJson = extractRenderJson(reply);
                    Map<String, Object> doneData = new HashMap<>();
                    doneData.put("content", reply);
                    if (renderJson != null) {
                        doneData.put("render", renderJson);
                    }
                    emitter.send(SseEmitter.event().name("done")
                        .data(objectMapper.writeValueAsString(doneData)));
                    emitter.complete();

                    aiUsageLogger.log("ADMIN", null, null, null,
                        System.currentTimeMillis() - startTime, true, null);
                } catch (Exception ignored) {}
            }
        );
        subscriptionRef.set(subscription);

        emitter.onCompletion(() -> subscriptionRef.get().dispose());
        emitter.onTimeout(() -> subscriptionRef.get().dispose());
        emitter.onError(e -> subscriptionRef.get().dispose());

        return emitter;
    }

    /**
     * 从 AI 回复中提取 render JSON（放在 ```json ... ``` 块中或行内）
     */
    private String extractRenderJson(String text) {
        try {
            int start = text.indexOf("{\"render\"");
            if (start == -1) return null;
            // 找到包含 render 的 JSON 块
            int braceCount = 0;
            boolean inString = false;
            int end = start;
            for (int i = start; i < text.length(); i++) {
                char c = text.charAt(i);
                if (c == '"' && (i == 0 || text.charAt(i - 1) != '\\')) {
                    inString = !inString;
                }
                if (!inString) {
                    if (c == '{') braceCount++;
                    else if (c == '}') braceCount--;
                    if (braceCount == 0) {
                        end = i + 1;
                        break;
                    }
                }
            }
            String json = text.substring(start, end);
            // 验证是合法 JSON
            objectMapper.readTree(json);
            return json;
        } catch (Exception e) {
            return null;
        }
    }
}