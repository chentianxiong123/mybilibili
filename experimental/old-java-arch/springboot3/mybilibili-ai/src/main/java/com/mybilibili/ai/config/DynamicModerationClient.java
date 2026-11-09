package com.mybilibili.ai.config;

import com.mybilibili.ai.entity.AiApiConfig;
import com.mybilibili.ai.entity.AiBinding;
import com.mybilibili.ai.mapper.AiApiConfigMapper;
import com.mybilibili.ai.mapper.AiBindingMapper;
import com.mybilibili.ai.service.ModerationProvider;
import com.mybilibili.ai.service.impl.XiaoShuiGuanModerationProvider;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import jakarta.annotation.PostConstruct;
import java.util.List;

/**
 * 内容审核提供者管理器。
 * 通过 ai_bindings 表获取功能绑定的渠道配置（feature='REVIEW'），
 * 加载对应提供者。
 */
@Slf4j
@Component
public class DynamicModerationClient {

    @Autowired
    private AiApiConfigMapper aiApiConfigMapper;

    @Autowired
    private AiBindingMapper aiBindingMapper;

    private ModerationProvider cachedProvider;

    @PostConstruct
    public void init() {
        refreshProviders();
    }

    /**
     * 获取内容审核提供者
     */
    public ModerationProvider getProvider() {
        return cachedProvider;
    }

    /**
     * 刷新提供者（根据绑定配置）
     */
    public void refreshProviders() {
        // 查找 REVIEW 功能绑定的渠道
        AiBinding binding = aiBindingMapper.selectByFeature("REVIEW");
        if (binding == null) {
            log.warn("未找到 REVIEW 功能绑定");
            cachedProvider = null;
            return;
        }

        AiApiConfig config = aiApiConfigMapper.selectById(binding.getApiConfigId());
        if (config == null || !config.getEnabled()) {
            log.warn("REVIEW 功能绑定的渠道不存在或已禁用");
            cachedProvider = null;
            return;
        }

        if ("小水管".equals(config.getName()) || config.getBaseUrl().contains("pie-xian")) {
            XiaoShuiGuanModerationProvider provider =
                    new XiaoShuiGuanModerationProvider(
                            config.getBaseUrl(),
                            config.getApiKey(),
                            config.getModel()
                    );
            cachedProvider = provider;
            log.info("内容审核提供者已加载: {} (model={})", config.getName(), config.getModel());
        } else {
            log.info("内容审核提供者: {}", config.getName());
        }
    }

    public void rebuild() {
        refreshProviders();
    }
}