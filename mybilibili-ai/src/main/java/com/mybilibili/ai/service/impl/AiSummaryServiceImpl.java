package com.mybilibili.ai.service.impl;

import com.mybilibili.ai.config.DeepSeekConfig;
import com.mybilibili.ai.service.AiConfigService;
import com.mybilibili.ai.service.AiSummaryService;
import com.mybilibili.ai.utils.SubtitleTextUtils;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.*;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestTemplate;

import java.io.BufferedWriter;
import java.io.File;
import java.io.FileWriter;
import java.io.IOException;
import java.text.SimpleDateFormat;
import java.util.*;

@Service
public class AiSummaryServiceImpl implements AiSummaryService {

    @Autowired
    private DeepSeekConfig deepSeekConfig;

    @Autowired
    private AiConfigService aiConfigService;

    private final RestTemplate restTemplate = new RestTemplate();

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
            return generateFallbackSummary(videoTitle, videoDescription);
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
        try {
            Map<String, Object> requestBody = new HashMap<>();
            requestBody.put("model", deepSeekConfig.getModel());

            List<Map<String, String>> messages = new ArrayList<>();

            Map<String, String> systemMessage = new HashMap<>();
            systemMessage.put("role", "system");
            systemMessage.put("content", SYSTEM_PROMPT);
            messages.add(systemMessage);

            Map<String, String> userMessage = new HashMap<>();
            userMessage.put("role", "user");
            userMessage.put("content", userPrompt);
            messages.add(userMessage);

            requestBody.put("messages", messages);
            requestBody.put("max_tokens", deepSeekConfig.getMaxTokens());
            requestBody.put("temperature", deepSeekConfig.getTemperature());

            HttpHeaders headers = new HttpHeaders();
            headers.setContentType(MediaType.APPLICATION_JSON);
            headers.set("Authorization", "Bearer " + aiConfigService.getApiKey());

            HttpEntity<Map<String, Object>> request = new HttpEntity<>(requestBody, headers);
            ResponseEntity<Map> response = restTemplate.postForEntity(deepSeekConfig.getApiUrl(), request, Map.class);

            if (response.getStatusCode() == HttpStatus.OK && response.getBody() != null) {
                return parseResponse(response.getBody());
            } else {
                return "API调用失败，状态码: " + response.getStatusCode();
            }

        } catch (Exception e) {
            return "API调用异常: " + e.getMessage();
        }
    }

    private String parseResponse(Map<String, Object> responseBody) {
        try {
            List<Map<String, Object>> choices = (List<Map<String, Object>>) responseBody.get("choices");
            if (choices != null && !choices.isEmpty()) {
                Map<String, Object> firstChoice = choices.get(0);
                Map<String, Object> message = (Map<String, Object>) firstChoice.get("message");
                if (message != null) {
                    return (String) message.get("content");
                }
            }
            return "无法解析API响应";
        } catch (Exception e) {
            return "解析响应异常: " + e.getMessage();
        }
    }

    private String generateFallbackSummary(String videoTitle, String videoDescription) {
        StringBuilder summary = new StringBuilder();
        summary.append("【视频摘要】\n");
        summary.append("视频标题: ").append(videoTitle).append("\n\n");

        if (videoDescription != null && !videoDescription.isEmpty()) {
            summary.append("视频描述: ").append(videoDescription).append("\n\n");
        }

        summary.append("该视频暂无字幕内容，无法生成详细摘要。\n");
        summary.append("请上传字幕文件或等待系统自动生成字幕后重试。\n\n");
        summary.append("【关键要点】\n");
        summary.append("1. 视频暂无字幕数据\n");
        summary.append("2. 需要字幕才能生成AI摘要\n");
        summary.append("3. 可使用字幕生成功能自动生成字幕");

        return summary.toString();
    }

    @Override
    public TestResult testApiConnection(String testText) {
        long startTime = System.currentTimeMillis();

        try {
            if (aiConfigService.getApiKey() == null || aiConfigService.getApiKey().isEmpty()) {
                return new TestResult(false, "API密钥未配置，请在application.yml中设置ai.deepseek.api-key");
            }

            Map<String, Object> requestBody = new HashMap<>();
            requestBody.put("model", deepSeekConfig.getModel());

            List<Map<String, String>> messages = new ArrayList<>();
            Map<String, String> message = new HashMap<>();
            message.put("role", "user");
            message.put("content", testText != null && !testText.isEmpty() ? testText : "你好，请回复'API测试成功'");
            messages.add(message);

            requestBody.put("messages", messages);
            requestBody.put("max_tokens", 100);

            HttpHeaders headers = new HttpHeaders();
            headers.setContentType(MediaType.APPLICATION_JSON);
            headers.set("Authorization", "Bearer " + aiConfigService.getApiKey());

            HttpEntity<Map<String, Object>> request = new HttpEntity<>(requestBody, headers);
            ResponseEntity<Map> response = restTemplate.postForEntity(deepSeekConfig.getApiUrl(), request, Map.class);

            long responseTime = System.currentTimeMillis() - startTime;

            if (response.getStatusCode() == HttpStatus.OK && response.getBody() != null) {
                String responseContent = parseResponse(response.getBody());
                return new TestResult(true, "API连接成功", responseContent, responseTime);
            } else {
                return new TestResult(false, "API返回错误，状态码: " + response.getStatusCode(), null, responseTime);
            }

        } catch (Exception e) {
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
            System.err.println("保存摘要文件失败: " + e.getMessage());
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