package com.mybilibili.search.entity;

import lombok.Data;
import org.springframework.data.annotation.Id;
import org.springframework.data.mongodb.core.mapping.Document;

import java.time.LocalDateTime;

@Data
@Document(collection = "recommend_config")
public class RecommendConfig {

    @Id
    private String id;

    private Double profileDecayFactor;
    private Double watchWeight;
    private Double likeWeight;
    private Double collectWeight;

    private Float categoryBoost;
    private Float tagBoost;
    private Integer topCategoryCount;
    private Integer topTagCount;

    private Integer hotFillMinCount;
    private Double freshnessBoost;
    private Integer freshnessWindowDays;

    private Integer viewDeduplicationMinutes;

    private Integer sentinelVideoQps;
    private Integer sentinelUserQps;
    private Integer sentinelSearchQps;
    private Integer sentinelAuthQps;

    private LocalDateTime updatedAt;
    private String updatedBy;

    public static RecommendConfig defaults() {
        RecommendConfig config = new RecommendConfig();
        config.setProfileDecayFactor(0.95);
        config.setWatchWeight(1.0);
        config.setLikeWeight(3.0);
        config.setCollectWeight(5.0);
        config.setCategoryBoost(2.0f);
        config.setTagBoost(1.5f);
        config.setTopCategoryCount(5);
        config.setTopTagCount(10);
        config.setHotFillMinCount(5);
        config.setFreshnessBoost(1.5);
        config.setFreshnessWindowDays(7);
        config.setViewDeduplicationMinutes(30);
        config.setSentinelVideoQps(200);
        config.setSentinelUserQps(100);
        config.setSentinelSearchQps(100);
        config.setSentinelAuthQps(10);
        config.setUpdatedAt(LocalDateTime.now());
        config.setUpdatedBy("system");
        return config;
    }
}
