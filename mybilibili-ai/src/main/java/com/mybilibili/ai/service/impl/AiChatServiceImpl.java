package com.mybilibili.ai.service.impl;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.mybilibili.ai.config.DynamicChatClient;
import com.mybilibili.ai.entity.AiChatMessage;
import com.mybilibili.ai.entity.AiConversation;
import com.mybilibili.ai.mapper.AiChatMessageMapper;
import com.mybilibili.ai.mapper.AiConversationMapper;
import com.mybilibili.ai.service.AiChatService;
import com.mybilibili.ai.util.AiUsageLogger;
import org.springframework.ai.chat.messages.AssistantMessage;
import org.springframework.ai.chat.messages.Message;
import org.springframework.ai.chat.messages.UserMessage;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;
import reactor.core.Disposable;
import reactor.core.publisher.Flux;

import java.util.*;
import java.util.concurrent.atomic.AtomicReference;

@Service
public class AiChatServiceImpl implements AiChatService {

    @Autowired
    private DynamicChatClient dynamicChatClient;

    @Autowired
    private AiConversationMapper aiConversationMapper;

    @Autowired
    private AiChatMessageMapper aiChatMessageMapper;

    @Autowired
    private AiUsageLogger aiUsageLogger;

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
        Long convId = conversationId;

        if (convId == null) {
            conversation = createConversation(userId);
            convId = conversation.getId();
            isNew = true;
        } else {
            conversation = aiConversationMapper.selectById(convId);
            if (conversation == null || !conversation.getUserId().equals(userId)) {
                SseEmitter err = new SseEmitter(0L);
                try {
                    err.send(SseEmitter.event().name("error").data("会话不存在或无权访问"));
                    err.complete();
                } catch (Exception ignored) {}
                return err;
            }
        }

        AiChatMessage userMsg = new AiChatMessage();
        userMsg.setConversationId(convId);
        userMsg.setRole("user");
        userMsg.setContent(content);
        aiChatMessageMapper.insert(userMsg);

        if (isNew) {
            String t = content.length() > 30 ? content.substring(0, 30) + "..." : content;
            conversation.setTitle(t);
            aiConversationMapper.updateById(conversation);
        }

        SseEmitter emitter = new SseEmitter(120000L);
        Long finalConvId = convId;
        AiConversation finalConv = conversation;

        // 流式调用
        List<Message> historyMessages = buildContextMessages(finalConvId, userMsg.getId(), content);
        long startTime = System.currentTimeMillis();

        Flux<String> flux = dynamicChatClient.getClient("CHAT").prompt()
                .system(SYSTEM_PROMPT)
                .messages(historyMessages)
                .stream()
                .content();

        StringBuilder fullResp = new StringBuilder();
        AtomicReference<Disposable> subscriptionRef = new AtomicReference<>();

        Disposable subscription = flux.subscribe(
            chunk -> {
                try {
                    fullResp.append(chunk);
                    emitter.send(SseEmitter.event().name("data").data(chunk));
                } catch (Exception ignored) {}
            },
            error -> {
                aiUsageLogger.log("CHAT", "deepseek-r1", null, null, System.currentTimeMillis() - startTime, false, error.getMessage());
                try {
                    emitter.send(SseEmitter.event().name("error").data("回复生成失败: " + error.getMessage()));
                    emitter.complete();
                } catch (Exception ignored) {}
            },
            () -> {
                try {
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

                    aiUsageLogger.log("CHAT", "deepseek-r1", null, null, System.currentTimeMillis() - startTime, true, null);

                    Map<String, Object> done = new HashMap<>();
                    done.put("conversationId", finalConvId);
                    emitter.send(SseEmitter.event().name("done").data(objectMapper.writeValueAsString(done)));
                    emitter.complete();
                } catch (Exception ignored) {}
            }
        );
        subscriptionRef.set(subscription);

        emitter.onCompletion(() -> subscriptionRef.get().dispose());
        emitter.onTimeout(() -> subscriptionRef.get().dispose());
        emitter.onError(e -> subscriptionRef.get().dispose());

        return emitter;
    }

    private List<Message> buildContextMessages(Long conversationId, Long excludeMsgId, String currentContent) {
        List<AiChatMessage> history = aiChatMessageMapper.selectByConversationId(conversationId);
        List<Message> messages = new ArrayList<>();

        int start = Math.max(0, history.size() - MAX_CONTEXT_MESSAGES);
        for (int i = start; i < history.size(); i++) {
            AiChatMessage m = history.get(i);
            if (m.getId().equals(excludeMsgId)) continue;
            if ("user".equals(m.getRole())) {
                messages.add(new UserMessage(m.getContent()));
            } else if ("assistant".equals(m.getRole())) {
                messages.add(new AssistantMessage(m.getContent()));
            }
        }

        messages.add(new UserMessage(currentContent));
        return messages;
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
