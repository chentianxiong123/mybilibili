package com.mybilibili.video.mapper;

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
            "COALESCE(SUM(danmaku_count), 0) as totalDanmaku, " +
            "COUNT(*) as totalManuscripts " +
            "FROM manuscripts WHERE user_id = #{userId} AND status = 3")
    Map<String, Object> selectTotalStatsByUserId(@Param("userId") Integer userId);

    @Select("SELECT COUNT(*) FROM user_interactions WHERE target_id = #{userId} AND target_type = 'USER' AND interaction_type = 'FOLLOW'")
    Integer selectFollowerCount(@Param("userId") Integer userId);

    @Select("SELECT " +
            "COALESCE(SUM(m.view_count), 0) as viewsIncrease, " +
            "COALESCE(SUM(m.like_count), 0) as likesIncrease, " +
            "COALESCE(SUM(m.comment_count), 0) as commentsIncrease, " +
            "COALESCE(SUM(m.danmaku_count), 0) as danmakuIncrease, " +
            "COALESCE(SUM(m.share_count), 0) as sharesIncrease, " +
            "COALESCE(SUM(m.collect_count), 0) as collectionsIncrease, " +
            "COALESCE(SUM(m.coin_count), 0) as coinsIncrease " +
            "FROM manuscripts m " +
            "WHERE m.user_id = #{userId} AND m.status = 3 AND DATE(m.upload_time) >= #{startDate}")
    Map<String, Object> selectIncreaseStats(@Param("userId") Integer userId, @Param("startDate") String startDate);

    @Select("SELECT " +
            "DATE(upload_time) as date, " +
            "SUM(view_count) as views, " +
            "SUM(like_count) as likes, " +
            "SUM(comment_count) as comments " +
            "FROM manuscripts " +
            "WHERE user_id = #{userId} AND status = 3 " +
            "AND DATE(upload_time) BETWEEN #{startDate} AND #{endDate} " +
            "GROUP BY DATE(upload_time)")
    List<Map<String, Object>> selectTrendData(@Param("userId") Integer userId, 
                                               @Param("startDate") String startDate, 
                                               @Param("endDate") String endDate);

    @Select("SELECT * FROM manuscripts WHERE user_id = #{userId} AND status = 3 ORDER BY view_count DESC LIMIT #{limit}")
    List<Manuscript> selectTopByViews(@Param("userId") Integer userId, @Param("limit") Integer limit);

    @Select("SELECT * FROM manuscripts WHERE user_id = #{userId} AND status = 3 ORDER BY like_count DESC LIMIT #{limit}")
    List<Manuscript> selectTopByLikes(@Param("userId") Integer userId, @Param("limit") Integer limit);

    @Select("SELECT * FROM manuscripts WHERE user_id = #{userId} AND status = 3 ORDER BY comment_count DESC LIMIT #{limit}")
    List<Manuscript> selectTopByComments(@Param("userId") Integer userId, @Param("limit") Integer limit);

    @Select("SELECT c.id, c.content, c.created_at as time, u.username, u.avatar, m.title as manuscriptTitle " +
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
}
