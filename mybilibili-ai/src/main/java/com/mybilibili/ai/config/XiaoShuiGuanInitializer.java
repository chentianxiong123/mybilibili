package com.mybilibili.ai.config;

import com.mybilibili.ai.entity.AiApiConfig;
import com.mybilibili.ai.entity.AiBinding;
import com.mybilibili.ai.mapper.AiApiConfigMapper;
import com.mybilibili.ai.mapper.AiBindingMapper;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import jakarta.annotation.PostConstruct;
import java.util.Date;
import java.util.List;

/**
 * 小水管审核渠道初始化器。
 * 启动时自动检查并添加小水管渠道和 REVIEW 功能绑定（如果不存在）。
 */
@Slf4j
@Component
public class XiaoShuiGuanInitializer {

    @Autowired
    private AiApiConfigMapper aiApiConfigMapper;

    @Autowired
    private AiBindingMapper aiBindingMapper;

    @Autowired
    private DynamicModerationClient moderationClient;

    private static final String PROVIDER_NAME = "小水管";
    private static final String BASE_URL = "https://api.pie-xian.com";
    private static final String MODEL = "qwen3guard";
    private static final String TYPE = "MODERATION";
    private static final String FEATURE = "REVIEW";

    @PostConstruct
    public void init() {
        // 检查 REVIEW 功能是否已绑定
        AiBinding existingBinding = aiBindingMapper.selectByFeature(FEATURE);
        if (existingBinding != null) {
            log.info("REVIEW 功能已绑定，跳过初始化");
            moderationClient.rebuild();
            return;
        }

        // 检查渠道是否已存在
        List<AiApiConfig> existingConfigs = aiApiConfigMapper.selectList(
                new com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper<AiApiConfig>()
                        .eq(AiApiConfig::getName, PROVIDER_NAME)
        );

        AiApiConfig config;
        if (existingConfigs.isEmpty()) {
            String apiKey = resolveApiKey();
            // 创建新渠道
            config = new AiApiConfig();
            config.setName(PROVIDER_NAME);
            config.setType(TYPE);
            config.setBaseUrl(BASE_URL);
            config.setApiKey(apiKey);
            config.setModel(MODEL);
            config.setMaxTokens(2000);
            config.setTemperature(0.7);
            config.setEnabled(true);
            config.setCreatedAt(new Date());
            config.setUpdatedAt(new Date());
            aiApiConfigMapper.insert(config);
            log.info("小水管渠道已创建: id={}", config.getId());
        } else {
            config = existingConfigs.get(0);
            log.info("小水管渠道已存在: id={}", config.getId());
        }

        // 绑定 REVIEW 功能
        AiBinding binding = new AiBinding();
        binding.setFeature(FEATURE);
        binding.setApiConfigId(config.getId());
        aiBindingMapper.insert(binding);
        log.info("REVIEW 功能绑定已创建: configId={}", config.getId());

        // 刷新提供者
        moderationClient.rebuild();
    }

    private String resolveApiKey() {
        String apiKey = firstNonBlank(
                System.getProperty("xiaoshuiguan.api-key"),
                System.getenv("XIAOSHUIGUAN_API_KEY")
        );
        if (apiKey == null || apiKey.isBlank()) {
            throw new IllegalStateException("小水管 API key is required: set -Dxiaoshuiguan.api-key or XIAOSHUIGUAN_API_KEY");
        }
        return apiKey;
    }

    private String firstNonBlank(String... values) {
        for (String value : values) {
            if (value != null && !value.isBlank()) {
                return value;
            }
        }
        return null;
    }
}
