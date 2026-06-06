package com.mybilibili.search.service;

import com.mybilibili.search.entity.RecommendConfig;
import com.mybilibili.search.repository.RecommendConfigRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.List;

@Service
public class RecommendConfigService {

    @Autowired
    private RecommendConfigRepository configRepository;

    private volatile RecommendConfig cachedConfig;

    public RecommendConfig getConfig() {
        if (cachedConfig != null) return cachedConfig;
        List<RecommendConfig> configs = configRepository.findAll();
        if (configs.isEmpty()) {
            RecommendConfig defaults = RecommendConfig.defaults();
            cachedConfig = configRepository.save(defaults);
        } else {
            cachedConfig = configs.get(0);
        }
        return cachedConfig;
    }

    public RecommendConfig updateConfig(RecommendConfig update, String adminName) {
        RecommendConfig current = getConfig();
        if (update.getProfileDecayFactor() != null) current.setProfileDecayFactor(update.getProfileDecayFactor());
        if (update.getWatchWeight() != null) current.setWatchWeight(update.getWatchWeight());
        if (update.getLikeWeight() != null) current.setLikeWeight(update.getLikeWeight());
        if (update.getCollectWeight() != null) current.setCollectWeight(update.getCollectWeight());
        if (update.getCategoryBoost() != null) current.setCategoryBoost(update.getCategoryBoost());
        if (update.getTagBoost() != null) current.setTagBoost(update.getTagBoost());
        if (update.getTopCategoryCount() != null) current.setTopCategoryCount(update.getTopCategoryCount());
        if (update.getTopTagCount() != null) current.setTopTagCount(update.getTopTagCount());
        if (update.getHotFillMinCount() != null) current.setHotFillMinCount(update.getHotFillMinCount());
        if (update.getFreshnessBoost() != null) current.setFreshnessBoost(update.getFreshnessBoost());
        if (update.getFreshnessWindowDays() != null) current.setFreshnessWindowDays(update.getFreshnessWindowDays());
        if (update.getViewDeduplicationMinutes() != null) current.setViewDeduplicationMinutes(update.getViewDeduplicationMinutes());
        if (update.getSentinelVideoQps() != null) current.setSentinelVideoQps(update.getSentinelVideoQps());
        if (update.getSentinelUserQps() != null) current.setSentinelUserQps(update.getSentinelUserQps());
        if (update.getSentinelSearchQps() != null) current.setSentinelSearchQps(update.getSentinelSearchQps());
        if (update.getSentinelAuthQps() != null) current.setSentinelAuthQps(update.getSentinelAuthQps());
        current.setUpdatedAt(LocalDateTime.now());
        current.setUpdatedBy(adminName);
        cachedConfig = configRepository.save(current);
        return cachedConfig;
    }

    public void invalidateCache() {
        cachedConfig = null;
    }
}
