package com.mybilibili.ai.util;

import com.mybilibili.ai.entity.AiUsageLog;
import com.mybilibili.ai.mapper.AiUsageLogMapper;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.util.Date;

@Slf4j
@Component
public class AiUsageLogger {

    @Autowired
    private AiUsageLogMapper aiUsageLogMapper;

    public void log(String feature, String model, Integer inputTokens,
                    Integer outputTokens, long durationMs, boolean success, String errorMessage) {
        try {
            AiUsageLog log = new AiUsageLog();
            log.setFeature(feature);
            log.setModel(model);
            log.setInputTokens(inputTokens != null ? inputTokens : 0);
            log.setOutputTokens(outputTokens != null ? outputTokens : 0);
            log.setTotalTokens(log.getInputTokens() + log.getOutputTokens());
            log.setDurationMs(durationMs);
            log.setStatus(success ? "SUCCESS" : "FAILED");
            log.setErrorMessage(errorMessage);
            log.setCreatedAt(new Date());
            aiUsageLogMapper.insert(log);
        } catch (Exception e) {
            log.error("写入AI用量日志失败", e);
        }
    }
}
