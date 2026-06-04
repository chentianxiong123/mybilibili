package com.mybilibili.analytics.mapper;

import org.apache.ibatis.annotations.Insert;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.time.LocalDateTime;

@Mapper
public interface ManuscriptAnalyticsMapper {

    @Insert("INSERT INTO manuscript_status_events " +
            "(manuscript_id, user_id, from_status, to_status, action, operator_type, operator_id, reason, created_at) " +
            "VALUES (#{manuscriptId}, #{userId}, #{fromStatus}, #{toStatus}, #{action}, #{operatorType}, #{operatorId}, #{reason}, #{occurredAt})")
    int insertStatusEvent(@Param("manuscriptId") Integer manuscriptId,
                          @Param("userId") Integer userId,
                          @Param("fromStatus") Integer fromStatus,
                          @Param("toStatus") Integer toStatus,
                          @Param("action") String action,
                          @Param("operatorType") String operatorType,
                          @Param("operatorId") Integer operatorId,
                          @Param("reason") String reason,
                          @Param("occurredAt") LocalDateTime occurredAt);

    @Insert("INSERT INTO manuscript_daily_metrics (metric_date, manuscript_id, user_id, view_count) " +
            "VALUES (DATE(#{occurredAt}), #{manuscriptId}, #{userId}, 1) " +
            "ON DUPLICATE KEY UPDATE view_count = view_count + 1, updated_at = NOW()")
    int incrementDailyView(@Param("manuscriptId") Integer manuscriptId,
                           @Param("userId") Integer userId,
                           @Param("occurredAt") LocalDateTime occurredAt);

    @Insert("INSERT INTO manuscript_daily_metrics (metric_date, manuscript_id, user_id, like_count) " +
            "VALUES (DATE(#{occurredAt}), #{manuscriptId}, #{userId}, GREATEST(#{delta}, 0)) " +
            "ON DUPLICATE KEY UPDATE like_count = GREATEST(like_count + #{delta}, 0), updated_at = NOW()")
    int incrementDailyLike(@Param("manuscriptId") Integer manuscriptId,
                           @Param("userId") Integer userId,
                           @Param("delta") Integer delta,
                           @Param("occurredAt") LocalDateTime occurredAt);

    @Insert("INSERT INTO manuscript_daily_metrics (metric_date, manuscript_id, user_id, coin_count) " +
            "VALUES (DATE(#{occurredAt}), #{manuscriptId}, #{userId}, GREATEST(#{delta}, 0)) " +
            "ON DUPLICATE KEY UPDATE coin_count = GREATEST(coin_count + #{delta}, 0), updated_at = NOW()")
    int incrementDailyCoin(@Param("manuscriptId") Integer manuscriptId,
                           @Param("userId") Integer userId,
                           @Param("delta") Integer delta,
                           @Param("occurredAt") LocalDateTime occurredAt);

    @Insert("INSERT INTO manuscript_daily_metrics (metric_date, manuscript_id, user_id, collect_count) " +
            "VALUES (DATE(#{occurredAt}), #{manuscriptId}, #{userId}, GREATEST(#{delta}, 0)) " +
            "ON DUPLICATE KEY UPDATE collect_count = GREATEST(collect_count + #{delta}, 0), updated_at = NOW()")
    int incrementDailyCollect(@Param("manuscriptId") Integer manuscriptId,
                              @Param("userId") Integer userId,
                              @Param("delta") Integer delta,
                              @Param("occurredAt") LocalDateTime occurredAt);

    @Insert("INSERT INTO manuscript_daily_metrics (metric_date, manuscript_id, user_id, share_count) " +
            "VALUES (DATE(#{occurredAt}), #{manuscriptId}, #{userId}, GREATEST(#{delta}, 0)) " +
            "ON DUPLICATE KEY UPDATE share_count = GREATEST(share_count + #{delta}, 0), updated_at = NOW()")
    int incrementDailyShare(@Param("manuscriptId") Integer manuscriptId,
                            @Param("userId") Integer userId,
                            @Param("delta") Integer delta,
                            @Param("occurredAt") LocalDateTime occurredAt);

    @Insert("INSERT INTO manuscript_daily_metrics (metric_date, manuscript_id, user_id, comment_count) " +
            "VALUES (DATE(#{occurredAt}), #{manuscriptId}, #{userId}, GREATEST(#{delta}, 0)) " +
            "ON DUPLICATE KEY UPDATE comment_count = GREATEST(comment_count + #{delta}, 0), updated_at = NOW()")
    int incrementDailyComment(@Param("manuscriptId") Integer manuscriptId,
                              @Param("userId") Integer userId,
                              @Param("delta") Integer delta,
                              @Param("occurredAt") LocalDateTime occurredAt);

    @Insert("INSERT INTO manuscript_daily_metrics (metric_date, manuscript_id, user_id, danmaku_count) " +
            "VALUES (DATE(#{occurredAt}), #{manuscriptId}, #{userId}, GREATEST(#{delta}, 0)) " +
            "ON DUPLICATE KEY UPDATE danmaku_count = GREATEST(danmaku_count + #{delta}, 0), updated_at = NOW()")
    int incrementDailyDanmaku(@Param("manuscriptId") Integer manuscriptId,
                              @Param("userId") Integer userId,
                              @Param("delta") Integer delta,
                              @Param("occurredAt") LocalDateTime occurredAt);
}
