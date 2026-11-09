package com.mybilibili.analytics.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.common.entity.Manuscript;
import org.apache.ibatis.annotations.*;

import java.util.List;
import java.util.Map;

@Mapper
public interface CreatorStatsMapper extends BaseMapper<Manuscript> {

    @Select("SELECT " +
            "COALESCE(SUM(view_count), 0) as totalViews, " +
            "COALESCE(SUM(like_count), 0) as totalLikes, " +
            "COALESCE(SUM(coin_count), 0) as totalCoins, " +
            "COALESCE(SUM(collect_count), 0) as totalCollections, " +
            "COALESCE(SUM(share_count), 0) as totalShares, " +
            "COALESCE(SUM(comment_count), 0) as totalComments, " +
            "COUNT(*) as totalManuscripts " +
            "FROM manuscripts WHERE user_id = #{userId} AND status = 3")
    Map<String, Object> selectTotalStatsByUserId(@Param("userId") Integer userId);

    @Select("SELECT COUNT(*) FROM user_interactions WHERE target_id = #{userId} AND target_type = 'USER' AND interaction_type = 'FOLLOW'")
    Integer selectFollowerCount(@Param("userId") Integer userId);

    @Select("SELECT " +
            "COALESCE(SUM(m.view_count), 0) as viewsIncrease, " +
            "COUNT(DISTINCT m.id) as manuscriptCount " +
            "FROM manuscripts m " +
            "WHERE m.user_id = #{userId} AND m.status = 3 AND DATE(m.upload_time) >= #{startDate}")
    Map<String, Object> selectIncreaseStats(@Param("userId") Integer userId, @Param("startDate") String startDate);

    @Select("SELECT COALESCE(SUM(view_count), 0) " +
            "FROM manuscript_daily_metrics " +
            "WHERE user_id = #{userId} AND metric_date >= #{startDate}")
    Integer selectViewIncreaseFromDailyMetrics(@Param("userId") Integer userId, @Param("startDate") String startDate);

    @Select("SELECT * FROM manuscripts WHERE user_id = #{userId} AND status = 3 ORDER BY view_count DESC LIMIT #{limit}")
    List<Manuscript> selectTopByViews(@Param("userId") Integer userId, @Param("limit") Integer limit);

    @Select("SELECT * FROM manuscripts WHERE user_id = #{userId} AND status = 3 ORDER BY like_count DESC LIMIT #{limit}")
    List<Manuscript> selectTopByLikes(@Param("userId") Integer userId, @Param("limit") Integer limit);

    @Select("SELECT * FROM manuscripts WHERE user_id = #{userId} AND status = 3 ORDER BY comment_count DESC LIMIT #{limit}")
    List<Manuscript> selectTopByComments(@Param("userId") Integer userId, @Param("limit") Integer limit);

    @Select("SELECT c.id, c.content, c.created_at as time, u.username, u.avatar, m.id as manuscriptId, m.title as manuscriptTitle " +
            "FROM comments c " +
            "JOIN users u ON c.user_id = u.id " +
            "JOIN manuscripts m ON c.manuscript_id = m.id " +
            "WHERE m.user_id = #{userId} AND m.status = 3 " +
            "ORDER BY c.created_at DESC " +
            "LIMIT #{limit}")
    List<Map<String, Object>> selectLatestComments(@Param("userId") Integer userId, @Param("limit") Integer limit);

    @Select("SELECT u.id, u.username, u.avatar, COUNT(*) as interactionCount " +
            "FROM user_interactions i " +
            "JOIN users u ON i.user_id = u.id " +
            "JOIN manuscripts m ON i.target_id = m.id " +
            "WHERE m.user_id = #{userId} AND m.status = 3 AND i.target_type = 'MANUSCRIPT' " +
            "GROUP BY u.id " +
            "ORDER BY interactionCount DESC " +
            "LIMIT #{limit}")
    List<Map<String, Object>> selectInteractionRanking(@Param("userId") Integer userId, @Param("limit") Integer limit);

    @Select("SELECT u.id, u.username, u.avatar, COUNT(DISTINCT c.id) as commentCount " +
            "FROM comments c " +
            "JOIN users u ON c.user_id = u.id " +
            "JOIN manuscripts m ON c.manuscript_id = m.id " +
            "WHERE m.user_id = #{userId} AND m.status = 3 " +
            "GROUP BY u.id " +
            "ORDER BY commentCount DESC " +
            "LIMIT #{limit}")
    List<Map<String, Object>> selectViewRanking(@Param("userId") Integer userId, @Param("limit") Integer limit);

    @Select("SELECT " +
            "COALESCE(SUM(view_count), 0) as viewCount, " +
            "COALESCE(SUM(like_count), 0) as likeCount, " +
            "COALESCE(SUM(comment_count), 0) as commentCount, " +
            "COALESCE(SUM(danmaku_count), 0) as danmakuCount, " +
            "COALESCE(SUM(coin_count), 0) as coinCount, " +
            "COALESCE(SUM(collect_count), 0) as collectCount, " +
            "COALESCE(SUM(share_count), 0) as shareCount, " +
            "COUNT(*) as manuscriptCount " +
            "FROM manuscripts WHERE user_id = #{userId} AND status = 3 AND DATE(upload_time) = #{date}")
    Map<String, Object> selectDailyManuscriptStats(@Param("userId") Integer userId, @Param("date") String date);

    @Select("SELECT " +
            "COALESCE(SUM(CASE WHEN interaction_type = 'FOLLOW' THEN 1 ELSE 0 END), 0) as newFollowers, " +
            "COALESCE(SUM(CASE WHEN interaction_type = 'UNFOLLOW' THEN 1 ELSE 0 END), 0) as unfollows " +
            "FROM user_interactions " +
            "WHERE target_id = #{userId} AND target_type = 'USER' AND DATE(created_at) = #{date}")
    Map<String, Object> selectDailyFanStats(@Param("userId") Integer userId, @Param("date") String date);

    @Select("SELECT COUNT(*) FROM user_interactions " +
            "WHERE target_id = #{userId} AND target_type = 'USER' AND interaction_type = 'FOLLOW' " +
            "AND DATE(created_at) <= #{date}")
    Integer selectFollowerCountAtDate(@Param("userId") Integer userId, @Param("date") String date);

    @Select("SELECT " +
            "COALESCE(SUM(view_count), 0) as totalViews, " +
            "COALESCE(SUM(like_count), 0) as totalLikes, " +
            "COALESCE(SUM(comment_count), 0) as totalComments, " +
            "COALESCE(SUM(danmaku_count), 0) as totalDanmaku, " +
            "COALESCE(SUM(coin_count), 0) as totalCoins, " +
            "COALESCE(SUM(collect_count), 0) as totalCollections, " +
            "COALESCE(SUM(share_count), 0) as totalShares, " +
            "COUNT(*) as totalManuscripts " +
            "FROM manuscripts WHERE user_id = #{userId} AND status = 3 AND DATE(upload_time) BETWEEN #{startDate} AND #{endDate}")
    Map<String, Object> selectOverviewByDateRange(@Param("userId") Integer userId, @Param("startDate") String startDate, @Param("endDate") String endDate);

    @Select("SELECT " +
            "DATE(created_at) as date, " +
            "SUM(CASE WHEN interaction_type = 'FOLLOW' THEN 1 ELSE 0 END) as newFollowers, " +
            "SUM(CASE WHEN interaction_type = 'UNFOLLOW' THEN 1 ELSE 0 END) as unfollows " +
            "FROM user_interactions " +
            "WHERE target_id = #{userId} AND target_type = 'USER' AND DATE(created_at) >= #{startDate} " +
            "GROUP BY DATE(created_at) " +
            "ORDER BY date ASC")
    List<Map<String, Object>> selectFansTrendData(@Param("userId") Integer userId, @Param("startDate") String startDate);

    @Select("SELECT " +
            "DATE(i.created_at) as date, " +
            "SUM(CASE WHEN i.interaction_type = 'LIKE' THEN 1 ELSE 0 END) as likes, " +
            "SUM(CASE WHEN i.interaction_type = 'COIN' THEN 1 ELSE 0 END) as coins, " +
            "SUM(CASE WHEN i.interaction_type = 'COLLECT' THEN 1 ELSE 0 END) as collects, " +
            "SUM(CASE WHEN i.interaction_type = 'SHARE' THEN 1 ELSE 0 END) as shares " +
            "FROM user_interactions i " +
            "JOIN manuscripts m ON i.target_id = m.id " +
            "WHERE m.user_id = #{userId} AND m.status = 3 " +
            "AND i.target_type = 'MANUSCRIPT' " +
            "AND DATE(i.created_at) >= #{startDate} " +
            "GROUP BY DATE(i.created_at) " +
            "ORDER BY date ASC")
    List<Map<String, Object>> selectInteractionTrendData(@Param("userId") Integer userId, @Param("startDate") String startDate);

    @Select("SELECT " +
            "metric_date as date, " +
            "SUM(view_count) as views " +
            "FROM manuscript_daily_metrics " +
            "WHERE user_id = #{userId} AND metric_date >= #{startDate} " +
            "GROUP BY metric_date " +
            "ORDER BY metric_date ASC")
    List<Map<String, Object>> selectViewTrendData(@Param("userId") Integer userId, @Param("startDate") String startDate);

    @Select("SELECT " +
            "DATE(c.created_at) as date, " +
            "COUNT(*) as comments " +
            "FROM comments c " +
            "JOIN manuscripts m ON c.manuscript_id = m.id " +
            "WHERE m.user_id = #{userId} AND m.status = 3 " +
            "AND DATE(c.created_at) >= #{startDate} " +
            "GROUP BY DATE(c.created_at) " +
            "ORDER BY date ASC")
    List<Map<String, Object>> selectCommentTrendData(@Param("userId") Integer userId, @Param("startDate") String startDate);

    @Select("SELECT id, title, view_count as views, upload_time " +
            "FROM manuscripts " +
            "WHERE user_id = #{userId} AND status = 3 " +
            "ORDER BY upload_time ASC")
    List<Map<String, Object>> selectManuscriptListForTrend(@Param("userId") Integer userId);

    @Select("SELECT " +
            "SUM(CASE WHEN i.interaction_type = 'LIKE' THEN 1 ELSE 0 END) as likesIncrease, " +
            "SUM(CASE WHEN i.interaction_type = 'COIN' THEN 1 ELSE 0 END) as coinsIncrease, " +
            "SUM(CASE WHEN i.interaction_type = 'COLLECT' THEN 1 ELSE 0 END) as collectionsIncrease, " +
            "SUM(CASE WHEN i.interaction_type = 'SHARE' THEN 1 ELSE 0 END) as sharesIncrease " +
            "FROM user_interactions i " +
            "JOIN manuscripts m ON i.target_id = m.id " +
            "WHERE m.user_id = #{userId} AND m.status = 3 " +
            "AND i.target_type = 'MANUSCRIPT' " +
            "AND DATE(i.created_at) >= #{startDate}")
    Map<String, Object> selectInteractionIncrease(@Param("userId") Integer userId, @Param("startDate") String startDate);

    @Select("SELECT COUNT(*) as commentsIncrease " +
            "FROM comments c JOIN manuscripts m ON c.manuscript_id = m.id " +
            "WHERE m.user_id = #{userId} AND m.status = 3 AND DATE(c.created_at) >= #{startDate}")
    Map<String, Object> selectCommentIncrease(@Param("userId") Integer userId, @Param("startDate") String startDate);
}
