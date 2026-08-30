package com.mybilibili.ai.service;

import com.mybilibili.ai.entity.AiApiConfig;
import java.util.List;

public interface AiApiConfigService {
    List<AiApiConfig> listAll();
    List<AiApiConfig> listByType(String type);
    AiApiConfig getById(Long id);
    AiApiConfig create(AiApiConfig config);
    AiApiConfig update(Long id, AiApiConfig config);
    void delete(Long id);
    void toggleEnabled(Long id);
    /** 根据 feature 获取绑定的渠道配置 */
    AiApiConfig getConfigForFeature(String feature);
    /** 设置功能绑定的渠道 */
    void bindFeature(String feature, Long configId);
    /** 获取所有功能绑定关系 */
    java.util.Map<String, Long> getAllBindings();

    /** 获取 Whisper 功能绑定配置 */
    java.util.Map<String, Object> getWhisperConfig();
}
