package com.mybilibili.ai.service.impl;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.mybilibili.ai.config.DynamicChatClient;
import com.mybilibili.ai.entity.AiApiConfig;
import com.mybilibili.ai.service.AiApiConfigService;
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
import java.util.Set;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.concurrent.atomic.AtomicReference;

@Service
public class AdminAiServiceImpl implements AdminAiService {

    @Autowired
    private DynamicChatClient dynamicChatClient;

    @Autowired
    private AdminToolService adminToolService;

    @Autowired
    private AiUsageLogger aiUsageLogger;

    @Autowired
    private AiApiConfigService aiApiConfigService;

    private final ObjectMapper objectMapper = new ObjectMapper();

    private static final String SYSTEM_PROMPT =
        "你是哔哩哔哩管理后台的 AI 助手，名叫\"管理助手\"。你的职责是帮助管理员完成以下工作：1. 查询平台统计数据，并以表格、图表或文字摘要形式呈现 2. 回答关于平台运营状况的问题。你可以使用以下工具获取数据：getOverviewStats: 获取平台概览统计；getUserGrowth(days): 获取用户增长趋势；getManuscriptStats: 获取稿件状态分布统计；getAiUsageOverview: 获取 AI 功能用量概览；getHotVideos(limit): 获取热门视频列表。当用户请求数据时，调用对应工具并将结果组织成清晰的中文回答。工具返回的 StatsData 中包含 chartType 字段。请在回答末尾用 JSON 格式返回渲染建议：{\"chartType\": \"number|line|pie|table\", \"title\": \"标题\", \"data\": {...}}";

    @Override
    public SseEmitter sendMessage(Integer adminId, String content) {
        SseEmitter emitter = new SseEmitter(120000L);
        long startTime = System.currentTimeMillis();

        ChatClient client = dynamicChatClient.getClient("ADMIN");
        String selectedModel = modelForFeature("ADMIN");
        if (client == null) {
            client = dynamicChatClient.getFirstActiveClient();
            selectedModel = firstActiveModel();
        }

        if (client == null) {
            try {
                emitter.send(SseEmitter.event().name("error").data("未配置可用的 AI 渠道"));
                emitter.complete();
            } catch (Exception ignored) {}
            return emitter;
        }
        final String model = selectedModel;

        List<ToolCallback> toolCallbacks = List.of(
            ToolCallbacks.from(adminToolService)
        );

        StringBuilder fullResp = new StringBuilder();
        AtomicReference<Disposable> subscriptionRef = new AtomicReference<>();
        AtomicBoolean completed = new AtomicBoolean(false);
        Set<String> calledTools = java.util.concurrent.ConcurrentHashMap.newKeySet();

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
                    // 检测工具调用关键字（AI 流式输出时会包含工具名提示）
                    for (String tool : new String[]{"getOverviewStats","getUserGrowth","getManuscriptStats",
                        "getAiUsageOverview","getHotVideos"}) {
                        if (chunk.contains(tool) && calledTools.add(tool)) {
                            emitter.send(SseEmitter.event().name("tool_call").data(tool));
                        }
                    }
                    emitter.send(SseEmitter.event().name("data").data(chunk));
                } catch (Exception e) {
                    disposeSubscription(subscriptionRef);
                    completeWithError(emitter, completed, e);
                }
            },
            error -> {
                aiUsageLogger.log("ADMIN", model, null, null,
                    System.currentTimeMillis() - startTime, false, error.getMessage());
                completeWithError(emitter, completed, error);
            },
            () -> {
                try {
                    String reply = fullResp.toString();
                    String renderJson = extractRenderJson(reply);
                    Map<String, Object> doneData = new HashMap<>();
                    doneData.put("content", reply);
                    if (renderJson != null) {
                        doneData.put("render", renderJson);
                    }
                    if (completed.compareAndSet(false, true)) {
                        emitter.send(SseEmitter.event().name("done")
                            .data(objectMapper.writeValueAsString(doneData)));
                        emitter.complete();
                    }

                    aiUsageLogger.log("ADMIN", model, null, null,
                        System.currentTimeMillis() - startTime, true, null);
                } catch (Exception e) {
                    disposeSubscription(subscriptionRef);
                    completeWithError(emitter, completed, e);
                }
            }
        );
        subscriptionRef.set(subscription);

        emitter.onCompletion(() -> disposeSubscription(subscriptionRef));
        emitter.onTimeout(() -> disposeSubscription(subscriptionRef));
        emitter.onError(e -> disposeSubscription(subscriptionRef));

        return emitter;
    }

    private void disposeSubscription(AtomicReference<Disposable> subscriptionRef) {
        Disposable subscription = subscriptionRef.getAndSet(null);
        if (subscription != null && !subscription.isDisposed()) {
            subscription.dispose();
        }
    }

    private void completeWithError(SseEmitter emitter, AtomicBoolean completed, Throwable error) {
        if (!completed.compareAndSet(false, true)) {
            return;
        }
        try {
            emitter.send(SseEmitter.event().name("error")
                    .data("请求失败: " + error.getMessage()));
        } catch (Exception ignored) {
        } finally {
            emitter.complete();
        }
    }

    /**
     * 从 AI 回复中提取 render JSON（放在 ```json ... ``` 块中或行内）
     */
    private String extractRenderJson(String text) {
        if (text == null || text.isEmpty()) {
            return null;
        }
        int searchFrom = 0;
        while (searchFrom < text.length()) {
            int start = text.indexOf('{', searchFrom);
            if (start == -1) {
                return null;
            }
            String json = extractJsonObject(text, start);
            if (json != null && isRenderJson(json)) {
                return json;
            }
            searchFrom = start + 1;
        }
        return null;
    }

    private String extractJsonObject(String text, int start) {
        int braceCount = 0;
        boolean inString = false;
        for (int i = start; i < text.length(); i++) {
            char c = text.charAt(i);
            if (c == '"' && (i == 0 || text.charAt(i - 1) != '\\')) {
                inString = !inString;
            }
            if (!inString) {
                if (c == '{') {
                    braceCount++;
                } else if (c == '}') {
                    braceCount--;
                    if (braceCount == 0) {
                        return text.substring(start, i + 1);
                    }
                }
            }
        }
        return null;
    }

    private boolean isRenderJson(String json) {
        try {
            var node = objectMapper.readTree(json);
            return node.has("render") || (node.has("chartType") && node.has("title") && node.has("data"));
        } catch (Exception e) {
            return false;
        }
    }

    private String modelForFeature(String feature) {
        AiApiConfig config = aiApiConfigService.getConfigForFeature(feature);
        return config != null ? config.getModel() : null;
    }

    private String firstActiveModel() {
        for (AiApiConfig config : aiApiConfigService.listAll()) {
            if (Boolean.TRUE.equals(config.getEnabled())) {
                return config.getModel();
            }
        }
        return null;
    }
}
