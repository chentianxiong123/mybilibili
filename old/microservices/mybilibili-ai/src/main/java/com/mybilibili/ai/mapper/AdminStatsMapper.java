package com.mybilibili.ai.mapper;

import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

import java.util.List;
import java.util.Map;

@Mapper
public interface AdminStatsMapper {

    @Select("""
            SELECT
                (SELECT COUNT(*) FROM users) AS totalUsers,
                (SELECT COUNT(*) FROM manuscripts) AS totalManuscripts,
                (SELECT COUNT(*) FROM manuscripts WHERE status = 3) AS publishedManuscripts,
                (SELECT COUNT(*) FROM manuscripts WHERE status = 0 OR review_status = 0) AS pendingReviews,
                (SELECT COUNT(*) FROM videos) AS totalVideos,
                (SELECT COUNT(*) FROM comments WHERE status = 0) AS totalComments,
                (SELECT COALESCE(SUM(view_count), 0) FROM manuscripts) AS totalViews,
                (SELECT COALESCE(SUM(like_count), 0) FROM manuscripts) AS totalLikes
            """)
    Map<String, Object> selectOverviewStats();

    @Select("""
            SELECT DATE(created_at) AS date, COUNT(*) AS count
            FROM users
            WHERE created_at >= #{startDate}
            GROUP BY DATE(created_at)
            ORDER BY date
            """)
    List<Map<String, Object>> selectUserGrowth(@Param("startDate") String startDate);

    @Select("""
            SELECT
                status,
                CASE status
                    WHEN -1 THEN '已下架'
                    WHEN 0 THEN '待审核'
                    WHEN 1 THEN '处理中'
                    WHEN 2 THEN '待发布'
                    WHEN 3 THEN '已发布'
                    WHEN 4 THEN '审核拒绝'
                    WHEN 5 THEN '处理失败'
                    ELSE CONCAT('状态', status)
                END AS statusText,
                COUNT(*) AS count
            FROM manuscripts
            GROUP BY status
            ORDER BY status
            """)
    List<Map<String, Object>> selectManuscriptStatusStats();

    @Select("""
            SELECT
                id,
                title,
                user_id AS userId,
                view_count AS viewCount,
                like_count AS likeCount,
                coin_count AS coinCount,
                collect_count AS collectCount,
                comment_count AS commentCount,
                danmaku_count AS danmakuCount,
                upload_time AS uploadTime
            FROM manuscripts
            WHERE status = 3
            ORDER BY view_count DESC, like_count DESC, id DESC
            LIMIT #{limit}
            """)
    List<Map<String, Object>> selectHotVideos(@Param("limit") int limit);
}
