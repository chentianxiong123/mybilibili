package com.mybilibili.ai.service.impl;

import com.mybilibili.ai.entity.AiApiConfig;
import com.mybilibili.ai.entity.AiBinding;
import com.mybilibili.ai.mapper.AiApiConfigMapper;
import com.mybilibili.ai.mapper.AiBindingMapper;
import com.mybilibili.ai.service.AiApiConfigService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import java.util.*;
import java.util.stream.Collectors;

@Service
public class AiApiConfigServiceImpl implements AiApiConfigService {

    @Autowired
    private AiApiConfigMapper aiApiConfigMapper;

    @Autowired
    private AiBindingMapper aiBindingMapper;

    @Override
    public List<AiApiConfig> listAll() {
        return aiApiConfigMapper.selectList(null);
    }

    @Override
    public List<AiApiConfig> listByType(String type) {
        return aiApiConfigMapper.selectList(
                new com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper<AiApiConfig>()
                        .eq(AiApiConfig::getType, type)
                        .eq(AiApiConfig::getEnabled, true)
        );
    }

    @Override
    public AiApiConfig getById(Long id) {
        return aiApiConfigMapper.selectById(id);
    }

    @Override
    public AiApiConfig create(AiApiConfig config) {
        config.setEnabled(true);
        config.setCreatedAt(new Date());
        config.setUpdatedAt(new Date());
        aiApiConfigMapper.insert(config);
        return config;
    }

    @Override
    public AiApiConfig update(Long id, AiApiConfig config) {
        config.setId(id);
        config.setUpdatedAt(new Date());
        aiApiConfigMapper.updateById(config);
        return config;
    }

    @Override
    public void delete(Long id) {
        // 先清除绑定关系
        aiBindingMapper.delete(new com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper<AiBinding>()
                .eq(AiBinding::getApiConfigId, id));
        aiApiConfigMapper.deleteById(id);
    }

    @Override
    public void toggleEnabled(Long id) {
        AiApiConfig config = aiApiConfigMapper.selectById(id);
        if (config != null) {
            config.setEnabled(!config.getEnabled());
            config.setUpdatedAt(new Date());
            aiApiConfigMapper.updateById(config);
        }
    }

    @Override
    public AiApiConfig getConfigForFeature(String feature) {
        AiBinding binding = aiBindingMapper.selectByFeature(feature);
        if (binding == null) return null;
        return aiApiConfigMapper.selectById(binding.getApiConfigId());
    }

    @Override
    public void bindFeature(String feature, Long configId) {
        AiBinding existing = aiBindingMapper.selectByFeature(feature);
        if (existing != null) {
            existing.setApiConfigId(configId);
            aiBindingMapper.updateById(existing);
        } else {
            AiBinding binding = new AiBinding();
            binding.setFeature(feature);
            binding.setApiConfigId(configId);
            aiBindingMapper.insert(binding);
        }
    }

    @Override
    public Map<String, Long> getAllBindings() {
        List<AiBinding> bindings = aiBindingMapper.selectList(null);
        return bindings.stream()
                .collect(Collectors.toMap(AiBinding::getFeature, AiBinding::getApiConfigId));
    }

}
