package com.mybilibili.ai.service.impl;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.mybilibili.ai.config.DeepSeekConfig;
import com.mybilibili.ai.entity.AiChatMessage;
import com.mybilibili.ai.entity.AiConversation;
import com.mybilibili.ai.mapper.AiChatMessageMapper;
import com.mybilibili.ai.mapper.AiConversationMapper;
import com.mybilibili.ai.service.AiChatService;
import com.mybilibili.ai.service.AiConfigService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.*;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestTemplate;
import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;
import java.util.*;

@Service
public class AiChatServiceImpl implements AiChatService {

    @Autowired
    private AiConfigService aiConfigService;

    @Autowired
    private DeepSeekConfig deepSeekConfig;

    @Autowired
    private AiConversationMapper aiConversationMapper;

    @Autowired
    private AiChatMessageMapper aiChatMessageMapper;

    private final RestTemplate restTemplate = new RestTemplate();
    private final ObjectMapper objectMapper = new ObjectMapper();

    private static final String SYSTEM_PROMPT =
        "你是一个哔哩哔哩平台的AI客服助手，你的名字叫\"哔哩助手\"。\n" +
        "你可以帮助用户解决以下问题：\n" +
        "1. 平台使用指南：如何投稿、如何互动、如何设置等\n" +
        "2. 账号相关问题：登录、注册、安全设置等\n" +
        "3. 内容相关问题：视频推荐、搜索技巧等\n" +
        "4. 其他与哔哩哔哩平台相关的问题\n\n" +
        "请在回答时保持友好、热情的语气，使用中文回答。如果你不知道某个问题的答案，" +
        "请诚实告知用户，不要编造信息。回答时使用清晰的结构，适当分段，让用户易于阅读。";

    private static final int MAX_CONTEXT_MESSAGES = 20;

    @Override
    public SseEmitter sendMessage(Integer userId, Long conversationId, String content) {
        AiConversation conversation;
        boolean isNew = false;

        if (conversationId == null) {
            conversation = createConversation(userId);
            conversationId = conversation.getId();
            isNew = true;
        } else {
            conversation = aiConversationMapper.selectById(conversationId);
            if (conversation == null || !conversation.getUserId().equals(userId)) {
                SseEmitter err = new SseEmitter(0L);
                try {
                    err.send(SseEmitter.event().name("error").data("会话不存在或无权访问"));
                    err.complete();
                } catch (Exception ignored) {}
                return err;
            }
        }

        // 保存用户消息
        AiChatMessage userMsg = new AiChatMessage();
        userMsg.setConversationId(conversationId);
        userMsg.setRole("user");
        userMsg.setContent(content);
        aiChatMessageMapper.insert(userMsg);

        // 首次对话自动生成标题
        if (isNew) {
            String t = content.length() > 30 ? content.substring(0, 30) + "..." : content;
            conversation.setTitle(t);
            aiConversationMapper.updateById(conversation);
        }

        SseEmitter emitter = new SseEmitter(120000L);
        Long finalConvId = conversationId;
        AiConversation finalConv = conversation;

        new Thread(() -> {
            try {
                emitter.send(SseEmitter.event().name("start").data(""));

                // 构建上下文
                List<AiChatMessage> history = aiChatMessageMapper.selectByConversationId(finalConvId);
                List<Map<String, String>> msgs = new ArrayList<>();

                Map<String, String> sys = new HashMap<>();
                sys.put("role", "system");
                sys.put("content", SYSTEM_PROMPT);
                msgs.add(sys);

                int start = Math.max(0, history.size() - MAX_CONTEXT_MESSAGES);
                for (int i = start; i < history.size(); i++) {
                    AiChatMessage m = history.get(i);
                    if (m.getId().equals(userMsg.getId())) continue;
                    Map<String, String> map = new HashMap<>();
                    map.put("role", m.getRole());
                    map.put("content", m.getContent());
                    msgs.add(map);
                }

                Map<String, String> cur = new HashMap<>();
                cur.put("role", "user");
                cur.put("content", content);
                msgs.add(cur);

                // 构建请求
                Map<String, Object> body = new HashMap<>();
                body.put("model", deepSeekConfig.getModel());
                body.put("messages", msgs);
                body.put("stream", true);
                body.put("max_tokens", deepSeekConfig.getMaxTokens());
                body.put("temperature", deepSeekConfig.getTemperature());

                String apiKey = aiConfigService.getApiKey();

                // 流式调用
                StringBuilder fullResp = new StringBuilder();

                restTemplate.execute(deepSeekConfig.getApiUrl(), HttpMethod.POST,
                    clientHttpRequest -> {
                        clientHttpRequest.getHeaders().setContentType(MediaType.APPLICATION_JSON);
                        clientHttpRequest.getHeaders().set("Authorization", "Bearer " + apiKey);
                        clientHttpRequest.getBody().write(objectMapper.writeValueAsBytes(body));
                    },
                    clientHttpResponse -> {
                        try (BufferedReader r = new BufferedReader(
                                new InputStreamReader(clientHttpResponse.getBody(), StandardCharsets.UTF_8))) {
                            String line;
                            while ((line = r.readLine()) != null) {
                                if (line.startsWith("data: ")) {
                                    String data = line.substring(6).trim();
                                    if ("[DONE]".equals(data)) break;
                                    try {
                                        Map<String, Object> j = objectMapper.readValue(data, Map.class);
                                        List<Map<String, Object>> choices = (List<Map<String, Object>>) j.get("choices");
                                        if (choices != null && !choices.isEmpty()) {
                                            Map<String, Object> delta = (Map<String, Object>) choices.get(0).get("delta");
                                            if (delta != null && delta.get("content") != null) {
                                                String chunk = (String) delta.get("content");
                                                fullResp.append(chunk);
                                                emitter.send(SseEmitter.event().name("data").data(chunk));
                                            }
                                        }
                                    } catch (Exception ignored) {}
                                }
                            }
                        }
                        return null;
                    }
                );

                // 保存 assistant 回复
                String reply = fullResp.toString();
                if (!reply.isEmpty()) {
                    AiChatMessage asst = new AiChatMessage();
                    asst.setConversationId(finalConvId);
                    asst.setRole("assistant");
                    asst.setContent(reply);
                    aiChatMessageMapper.insert(asst);
                }

                finalConv.setUpdatedAt(new Date());
                aiConversationMapper.updateById(finalConv);

                Map<String, Object> done = new HashMap<>();
                done.put("conversationId", finalConvId);
                emitter.send(SseEmitter.event().name("done").data(objectMapper.writeValueAsString(done)));
                emitter.complete();

            } catch (Exception e) {
                try {
                    emitter.send(SseEmitter.event().name("error").data("回复生成失败: " + e.getMessage()));
                    emitter.complete();
                } catch (Exception ignored) {}
            }
        }).start();

        return emitter;
    }

    @Override
    public AiConversation createConversation(Integer userId) {
        AiConversation c = new AiConversation();
        c.setUserId(userId);
        c.setTitle("AI客服对话");
        c.setStatus(1);
        c.setCreatedAt(new Date());
        c.setUpdatedAt(new Date());
        aiConversationMapper.insert(c);
        return c;
    }

    @Override
    public List<AiConversation> getConversations(Integer userId) {
        return aiConversationMapper.selectByUserId(userId);
    }

    @Override
    public List<AiChatMessage> getMessages(Long conversationId, Integer userId) {
        AiConversation c = aiConversationMapper.selectById(conversationId);
        if (c == null || !c.getUserId().equals(userId)) return Collections.emptyList();
        return aiChatMessageMapper.selectByConversationId(conversationId);
    }

    @Override
    public void deleteConversation(Long conversationId, Integer userId) {
        AiConversation c = aiConversationMapper.selectById(conversationId);
        if (c != null && c.getUserId().equals(userId)) {
            aiConversationMapper.deleteById(conversationId);
        }
    }
}