package com.mybilibili.interaction.service.impl;

import com.mybilibili.interaction.entity.UserProfile;
import com.mybilibili.interaction.repository.UserProfileRepository;
import com.mybilibili.interaction.service.UserProfileService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Service
public class UserProfileServiceImpl implements UserProfileService {

    private static final double WATCH_WEIGHT = 1.0;
    private static final double LIKE_WEIGHT = 3.0;
    private static final double COLLECT_WEIGHT = 5.0;
    private static final double DECAY_FACTOR = 0.95;
    private static final int MAX_TAGS = 50;

    @Autowired
    private UserProfileRepository userProfileRepository;

    @Override
    public UserProfile getOrCreateProfile(Integer userId) {
        return userProfileRepository.findByUserId(userId)
                .orElseGet(() -> {
                    UserProfile profile = UserProfile.createDefault(userId);
                    return userProfileRepository.save(profile);
                });
    }

    @Override
    public void recordWatch(Integer userId, Integer categoryId, List<String> tags, Integer durationSeconds) {
        UserProfile profile = getOrCreateProfile(userId);
        profile.setTotalWatchCount(profile.getTotalWatchCount() + 1);

        updateCategoryWeight(profile, categoryId, WATCH_WEIGHT);
        updateTagWeights(profile, tags, WATCH_WEIGHT);
        updateActiveHour(profile);
        updatePreferredDuration(profile, durationSeconds);

        profile.setLastUpdated(LocalDateTime.now());
        userProfileRepository.save(profile);
    }

    @Override
    public void recordLike(Integer userId, Integer categoryId, List<String> tags) {
        UserProfile profile = getOrCreateProfile(userId);
        profile.setTotalLikeCount(profile.getTotalLikeCount() + 1);

        updateCategoryWeight(profile, categoryId, LIKE_WEIGHT);
        updateTagWeights(profile, tags, LIKE_WEIGHT);

        profile.setLastUpdated(LocalDateTime.now());
        userProfileRepository.save(profile);
    }

    @Override
    public void recordCollect(Integer userId, Integer categoryId, List<String> tags) {
        UserProfile profile = getOrCreateProfile(userId);
        profile.setTotalCollectCount(profile.getTotalCollectCount() + 1);

        updateCategoryWeight(profile, categoryId, COLLECT_WEIGHT);
        updateTagWeights(profile, tags, COLLECT_WEIGHT);

        profile.setLastUpdated(LocalDateTime.now());
        userProfileRepository.save(profile);
    }

    private void updateCategoryWeight(UserProfile profile, Integer categoryId, double weight) {
        if (categoryId == null) return;
        Map<String, Double> weights = profile.getCategoryWeights();
        if (weights == null) weights = new HashMap<>();
        String key = categoryId.toString();
        weights.merge(key, weight, Double::sum);
        applyDecay(weights);
        profile.setCategoryWeights(weights);
    }

    private void updateTagWeights(UserProfile profile, List<String> tags, double weight) {
        if (tags == null || tags.isEmpty()) return;
        Map<String, Double> weights = profile.getTagWeights();
        if (weights == null) weights = new HashMap<>();
        for (String tag : tags) {
            if (tag != null && !tag.isBlank()) {
                weights.merge(tag.trim(), weight, Double::sum);
            }
        }
        applyDecay(weights);
        trimToMax(weights, MAX_TAGS);
        profile.setTagWeights(weights);
    }

    private void updateActiveHour(UserProfile profile) {
        Map<String, Integer> hours = profile.getActiveHours();
        if (hours == null) hours = new HashMap<>();
        String hour = String.valueOf(LocalDateTime.now().getHour());
        hours.merge(hour, 1, Integer::sum);
        profile.setActiveHours(hours);
    }

    private void updatePreferredDuration(UserProfile profile, Integer durationSeconds) {
        if (durationSeconds == null) return;
        if (durationSeconds < 300) {
            profile.setPreferredDuration("short");
        } else if (durationSeconds > 1200) {
            profile.setPreferredDuration("long");
        } else {
            profile.setPreferredDuration("medium");
        }
    }

    private void applyDecay(Map<String, Double> weights) {
        weights.replaceAll((k, v) -> v * DECAY_FACTOR);
        weights.entrySet().removeIf(e -> e.getValue() < 0.01);
    }

    private void trimToMax(Map<String, Double> weights, int max) {
        if (weights.size() <= max) return;
        weights.entrySet().stream()
                .sorted(Map.Entry.<String, Double>comparingByValue())
                .limit(weights.size() - max)
                .map(Map.Entry::getKey)
                .toList()
                .forEach(weights::remove);
    }
}
