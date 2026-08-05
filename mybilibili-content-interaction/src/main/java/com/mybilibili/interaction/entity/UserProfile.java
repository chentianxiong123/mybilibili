package com.mybilibili.interaction.entity;

import lombok.Data;
import org.springframework.data.annotation.Id;
import org.springframework.data.mongodb.core.index.Indexed;
import org.springframework.data.mongodb.core.mapping.Document;

import java.time.LocalDateTime;
import java.util.HashMap;
import java.util.Map;

@Data
@Document(collection = "user_profiles")
public class UserProfile {

    @Id
    private String id;

    @Indexed(unique = true)
    private Integer userId;

    private Map<String, Double> categoryWeights;

    private Map<String, Double> tagWeights;

    private String preferredDuration;

    private Map<String, Integer> activeHours;

    private Integer totalWatchCount;
    private Integer totalLikeCount;
    private Integer totalCollectCount;

    private LocalDateTime lastUpdated;
    private LocalDateTime createdAt;

    public static UserProfile createDefault(Integer userId) {
        UserProfile profile = new UserProfile();
        profile.setUserId(userId);
        profile.setCategoryWeights(new HashMap<>());
        profile.setTagWeights(new HashMap<>());
        profile.setActiveHours(new HashMap<>());
        profile.setPreferredDuration("medium");
        profile.setTotalWatchCount(0);
        profile.setTotalLikeCount(0);
        profile.setTotalCollectCount(0);
        profile.setCreatedAt(LocalDateTime.now());
        profile.setLastUpdated(LocalDateTime.now());
        return profile;
    }
}
