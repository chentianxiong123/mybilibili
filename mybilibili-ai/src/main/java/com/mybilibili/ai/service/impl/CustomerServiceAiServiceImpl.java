package com.mybilibili.ai.service.impl;

import com.mybilibili.ai.config.DynamicChatClient;
import com.mybilibili.ai.entity.AiSkill;
import com.mybilibili.ai.entity.AiSession;
import com.mybilibili.ai.entity.AiChatMessage;
import com.mybilibili.ai.mapper.AiSessionMapper;
import com.mybilibili.ai.mapper.AiChatMessageMapper;
import com.mybilibili.ai.mapper.ManuscriptMapper;
import com.mybilibili.ai.mapper.VideoMapper;
import com.mybilibili.ai.service.AiSkillService;
import com.mybilibili.ai.service.CustomerServiceAiService;
import com.mybilibili.ai.service.SkillRoutingService;
import com.mybilibili.ai.tool.CustomerServiceReadonlyToolSet;
import com.mybilibili.ai.util.AiUsageLogger;
import org.springframework.ai.support.ToolCallbacks;
import org.springframework.ai.tool.ToolCallback;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;

import java.util.Date;
import java.util.List;
import java.util.concurrent.atomic.AtomicReference;

@Service
public class CustomerServiceAiServiceImpl implements CustomerServiceAiService {

    public static final String TYPE_CUSTOMER_SERVICE = "CUSTOMER_SERVICE";
    public static final Integer STATUS_WAITING_HUMAN = 1;
    public static final Integer STATUS_ACTIVE = 0;
    public static final String TRANSFER_MARKER = "[TRANSFER_TO_HUMAN]";

    /** Base system prompt for customer service AI */
    public static final String BASE_SYSTEM_PROMPT = "你是哔哩哔哩平台的AI客服助手。请专业、友善地回答用户问题。\n" +
            "回答时保持清晰结构，必要时引导转人工服务。\n" +
            "你可以使用只读查询工具帮助用户查询其本人可见的稿件状态、视频处理状态、字幕/摘要生成状态和客服会话状态。\n" +
            "你不能承诺或执行任何写操作，包括审核、发布、下架、重试处理、转码、重新生成字幕或摘要；用户需要这些操作时应转人工服务。\n" +
            "当你判断需要转人工服务时，请在回复末尾添加转人工标记：[TRANSFER_TO_HUMAN]\n" +
            "转人工的条件包括：用户明确要求转人工、涉及复杂投诉、需要人工授权、情绪激动的用户等。";

    @Autowired
    private DynamicChatClient dynamicChatClient;

    @Autowired
    private AiUsageLogger aiUsageLogger;

    @Autowired
    private AiSessionMapper aiSessionMapper;

    @Autowired
    private AiChatMessageMapper aiChatMessageMapper;

    @Autowired
    private VideoMapper videoMapper;

    @Autowired
    private ManuscriptMapper manuscriptMapper;

    @Autowired
    private SkillRoutingService skillRoutingService;

    @Override
    public SseEmitter chat(Long userId, String content) {
        // 1. Get ChatClient for CHAT feature
        org.springframework.ai.chat.client.ChatClient client = dynamicChatClient.getClient("CHAT");
        if (client == null) {
            SseEmitter emitter = new SseEmitter(0L);
            try {
                emitter.send(SseEmitter.event().name("error").data("AI渠道未配置，请先去 AI 渠道管理 中绑定 CHAT 功能"));
                emitter.complete();
            } catch (Exception ignored) {}
            return emitter;
        }

        // 2. Get or create session
        AiSession session = getOrCreateSession(userId);

        // 3. Insert user message
        AiChatMessage userMsg = new AiChatMessage();
        userMsg.setSessionId(session.getId());
        userMsg.setRole("USER");
        userMsg.setContent(content);
        userMsg.setCreatedAt(new Date());
        aiChatMessageMapper.insert(userMsg);

        // 4. Route skills using LLM to determine which skill(s) to use
        List<AiSkill> matchedSkills = skillRoutingService.routeSkills(content);

        // 5. Build system prompt from matched skills or use base prompt only
        String systemPrompt = buildSystemPrompt(matchedSkills);
        List<ToolCallback> toolCallbacks = List.of(
                ToolCallbacks.from(new CustomerServiceReadonlyToolSet(userId, videoMapper, manuscriptMapper, aiSessionMapper))
        );

        // 6. 调用 client.prompt().system(SYSTEM_PROMPT).user(content).stream() 返回 SSE
        SseEmitter emitter = new SseEmitter(120000L);
        long startTime = System.currentTimeMillis();

        AtomicReference<reactor.core.Disposable> subscriptionRef = new AtomicReference<>();

        reactor.core.publisher.Flux<String> flux = client.prompt()
                .system(systemPrompt)
                .user(content)
                .tools(toolCallbacks)
                .stream()
                .content();

        StringBuilder fullResp = new StringBuilder();
        AtomicReference<Boolean> transferred = new AtomicReference<>(false);

        reactor.core.Disposable subscription = flux.subscribe(
                chunk -> {
                    try {
                        fullResp.append(chunk);
                        emitter.send(SseEmitter.event().name("data").data(chunk));
                    } catch (Exception ignored) {}
                },
                error -> {
                    aiUsageLogger.log("CHAT", "customer-service", null, null, System.currentTimeMillis() - startTime, false, error.getMessage());
                    try {
                        emitter.send(SseEmitter.event().name("error").data("回复生成失败: " + error.getMessage()));
                        emitter.complete();
                    } catch (Exception ignored) {}
                },
                () -> {
                    try {
                        String reply = fullResp.toString();
                        aiUsageLogger.log("CHAT", "customer-service", null, null, System.currentTimeMillis() - startTime, true, null);

                        // 检查是否需要转人工
                        if (reply.contains(TRANSFER_MARKER)) {
                            // 去掉转人工标记
                            reply = reply.replace(TRANSFER_MARKER, "").trim();
                            // 标记已转人工
                            transferred.set(true);
                        }

                        if (!reply.isEmpty()) {
                            // 插入 AI 助手消息
                            AiChatMessage asstMsg = new AiChatMessage();
                            asstMsg.setSessionId(session.getId());
                            asstMsg.setRole("ASSISTANT");
                            asstMsg.setContent(reply);
                            asstMsg.setCreatedAt(new Date());
                            aiChatMessageMapper.insert(asstMsg);
                        }

                        // 如果需要转人工
                        if (transferred.get()) {
                            // 插入系统消息
                            AiChatMessage sysMsg = new AiChatMessage();
                            sysMsg.setSessionId(session.getId());
                            sysMsg.setRole("SYSTEM");
                            sysMsg.setContent("[已转人工服务，请稍候]");
                            sysMsg.setCreatedAt(new Date());
                            aiChatMessageMapper.insert(sysMsg);

                            // 更新会话状态为等待人工
                            session.setStatus(STATUS_WAITING_HUMAN);
                            session.setUpdatedAt(new Date());
                            aiSessionMapper.updateById(session);

                            // 返回转人工标记
                            emitter.send(SseEmitter.event().name("transfer").data(""));
                        }

                        emitter.send(SseEmitter.event().name("done").data(""));
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

    @Override
    @Transactional
    public void transferToHuman(Long sessionId, Long userId, String reason) {
        // 1. 获取会话
        AiSession session = null;
        if (sessionId != null) {
            session = aiSessionMapper.selectById(sessionId);
        }
        if (session == null && userId != null) {
            // 按 userId 查找活跃会话
            try {
                var sessions = aiSessionMapper.selectByTypeAndStatus(TYPE_CUSTOMER_SERVICE, STATUS_ACTIVE);
                if (sessions != null && !sessions.isEmpty()) {
                    for (AiSession s : sessions) {
                        if (s.getUserId().equals(userId)) {
                            session = s;
                            break;
                        }
                    }
                }
            } catch (Exception e) {
                // Ignore
            }
        }

        if (session == null) {
            throw new RuntimeException("未找到会话，无法转人工");
        }

        // 2. 插入 SYSTEM 消息
        AiChatMessage sysMsg = new AiChatMessage();
        sysMsg.setSessionId(session.getId());
        sysMsg.setRole("SYSTEM");
        sysMsg.setContent("[已转人工服务，请稍候]");
        sysMsg.setCreatedAt(new Date());
        aiChatMessageMapper.insert(sysMsg);

        // 3. 更新会话状态为等待人工
        session.setStatus(STATUS_WAITING_HUMAN);
        session.setUpdatedAt(new Date());
        aiSessionMapper.updateById(session);
    }

    private AiSession getOrCreateSession(Long userId) {
        // 查找该用户的活跃客服会话
        AiSession session = null;
        try {
            var sessions = aiSessionMapper.selectByTypeAndStatus(TYPE_CUSTOMER_SERVICE, STATUS_ACTIVE);
            if (sessions != null && !sessions.isEmpty()) {
                for (AiSession s : sessions) {
                    if (s.getUserId().equals(userId)) {
                        session = s;
                        break;
                    }
                }
            }
        } catch (Exception e) {
            // Ignore and create new session
        }

        if (session == null) {
            session = new AiSession();
            session.setUserId(userId);
            session.setType(TYPE_CUSTOMER_SERVICE);
            session.setStatus(STATUS_ACTIVE);
            session.setCreatedAt(new Date());
            session.setUpdatedAt(new Date());
            aiSessionMapper.insert(session);
        }

        return session;
    }

    private String buildSystemPrompt(List<AiSkill> matchedSkills) {
        StringBuilder prompt = new StringBuilder();
        prompt.append(BASE_SYSTEM_PROMPT);

        // Add skill-specific prompts if any skills matched
        if (matchedSkills != null && !matchedSkills.isEmpty()) {
            for (AiSkill skill : matchedSkills) {
                if (skill.getSystemPrompt() != null && !skill.getSystemPrompt().isBlank()) {
                    prompt.append("\n\n").append(skill.getSystemPrompt());
                }
            }
            for (AiSkill skill : matchedSkills) {
                if (skill.getFewShotExamples() != null && !skill.getFewShotExamples().isBlank()) {
                    prompt.append("\n\n参考示例：\n").append(skill.getFewShotExamples());
                }
            }
        }

        return prompt.toString();
    }
}
