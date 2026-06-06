package com.mybilibili.ai.service.impl;

import org.springframework.ai.chat.client.ChatClient;
import com.mybilibili.ai.config.DynamicChatClient;
import com.mybilibili.ai.service.AiSummaryService;
import com.mybilibili.ai.utils.SubtitleTextUtils;
import com.mybilibili.ai.util.AiUsageLogger;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.io.BufferedWriter;
import java.io.File;
import java.io.FileWriter;
import java.io.IOException;
import java.text.SimpleDateFormat;
import java.util.Date;

@Slf4j
@Service
public class AiSummaryServiceImpl implements AiSummaryService {

    @Autowired
    private DynamicChatClient dynamicChatClient;

    @Autowired
    private AiUsageLogger aiUsageLogger;

    private static final String SYSTEM_PROMPT =
        "你是一个专业的视频内容分析助手。你的任务是根据视频字幕内容生成简洁、准确的视频摘要。\n\n" +
        "请遵循以下要求：\n" +
        "1. 摘要应包含视频的主要内容和核心观点\n" +
        "2. 使用简洁明了的语言，避免冗长\n" +
        "3. 突出视频的重点和亮点\n" +
        "4. 摘要长度控制在200-500字之间\n" +
        "5. 如果字幕内容不足以生成摘要，请基于已有信息给出简要说明\n\n" +
        "输出格式：\n" +
        "【视频摘要】\n" +
        "（在这里写摘要内容）\n\n" +
        "【关键要点】\n" +
        "1. 要点一\n" +
        "2. 要点二\n" +
        "3. 要点三";

    @Override
    public String generateSummary(String subtitleText, String videoTitle) {
        return generateSummary(subtitleText, videoTitle, null);
    }

    @Override
    public String generateSummary(String subtitleText, String videoTitle, String videoDescription) {
        if (subtitleText == null || subtitleText.trim().isEmpty()) {
            log.warn("AI摘要生成中止：字幕内容为空, title={}", videoTitle);
            return null;
        }

        String cleanedText = SubtitleTextUtils.cleanText(subtitleText);
        String truncatedText = SubtitleTextUtils.truncateText(cleanedText);
        String userPrompt = buildUserPrompt(truncatedText, videoTitle, videoDescription);

        return callAiApi(userPrompt);
    }

    private String buildUserPrompt(String subtitleText, String videoTitle, String videoDescription) {
        StringBuilder prompt = new StringBuilder();
        prompt.append("视频标题: ").append(videoTitle).append("\n\n");
        if (videoDescription != null && !videoDescription.isEmpty()) {
            prompt.append("视频描述: ").append(videoDescription).append("\n\n");
        }
        prompt.append("视频字幕内容:\n");
        prompt.append(subtitleText);
        return prompt.toString();
    }

    private String callAiApi(String userPrompt) {
        long start = System.currentTimeMillis();
        try {
            ChatClient client = dynamicChatClient.getClient("SUMMARY");
            if (client == null) {
                String message = "SUMMARY 渠道未配置或已禁用";
                aiUsageLogger.log("SUMMARY", null, null, null, System.currentTimeMillis() - start, false, message);
                log.warn("AI摘要生成中止：{}", message);
                return null;
            }

            String result = client.prompt()
                    .system(SYSTEM_PROMPT)
                    .user(userPrompt)
                    .call()
                    .content();
            aiUsageLogger.log("SUMMARY", "deepseek-r1", null, null, System.currentTimeMillis() - start, true, null);
            return result;
        } catch (Exception e) {
            aiUsageLogger.log("SUMMARY", "deepseek-r1", null, null, System.currentTimeMillis() - start, false, e.getMessage());
            log.error("AI摘要生成失败: {}", e.getMessage(), e);
            return null;
        }
    }

    @Override
    public TestResult testApiConnection(String testText) {
        long startTime = System.currentTimeMillis();
        try {
            String prompt = testText != null && !testText.isEmpty() ? testText : "你好，请回复'API测试成功'";
            ChatClient client = dynamicChatClient.getClient("CHAT");
            if (client == null) {
                client = dynamicChatClient.getFirstActiveClient();
            }
            if (client == null) {
                return new TestResult(false, "没有可用的 API 渠道，请先在渠道管理页面配置并启用渠道");
            }
            String responseContent = client.prompt()
                    .user(prompt)
                    .call()
                    .content();

            aiUsageLogger.log("TEST", "deepseek-r1", null, null, System.currentTimeMillis() - startTime, responseContent != null, null);

            long responseTime = System.currentTimeMillis() - startTime;
            return new TestResult(true, "API连接成功", responseContent, responseTime);

        } catch (Exception e) {
            aiUsageLogger.log("TEST", "deepseek-r1", null, null, System.currentTimeMillis() - startTime, false, e.getMessage());
            long responseTime = System.currentTimeMillis() - startTime;
            return new TestResult(false, "API调用异常: " + e.getMessage(), null, responseTime);
        }
    }

    @Override
    public boolean saveSummaryToFile(String summary, String filePath, String videoTitle) {
        try {
            File file = new File(filePath);
            File parentDir = file.getParentFile();
            if (parentDir != null && !parentDir.exists()) {
                parentDir.mkdirs();
            }

            try (BufferedWriter writer = new BufferedWriter(new FileWriter(file))) {
                SimpleDateFormat sdf = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss");
                writer.write(repeatChar('=', 60));
                writer.newLine();
                writer.write("视频标题: " + videoTitle);
                writer.newLine();
                writer.write("生成时间: " + sdf.format(new Date()));
                writer.newLine();
                writer.write(repeatChar('=', 60));
                writer.newLine();
                writer.newLine();
                writer.write(summary);
                writer.newLine();
            }
            return true;
        } catch (IOException e) {
            return false;
        }
    }

    private String repeatChar(char c, int count) {
        StringBuilder sb = new StringBuilder(count);
        for (int i = 0; i < count; i++) {
            sb.append(c);
        }
        return sb.toString();
    }
}
