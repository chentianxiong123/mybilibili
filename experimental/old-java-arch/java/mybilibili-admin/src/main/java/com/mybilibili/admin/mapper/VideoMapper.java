package com.mybilibili.admin.mapper;

import com.mybilibili.common.entity.Video;
import org.apache.ibatis.annotations.*;

import java.util.List;
import java.util.Map;

@Mapper
public interface VideoMapper {
    @Select("<script>" +
            "SELECT * FROM videos " +
            "<where>" +
            "  <if test='keyword != null and keyword != \"\"'>" +
            "    AND (title LIKE CONCAT('%', #{keyword}, '%') OR description LIKE CONCAT('%', #{keyword}, '%'))" +
            "  </if>" +
            "  <if test='status != null'>" +
            "    AND status = #{status}" +
            "  </if>" +
            "</where>" +
            "ORDER BY id DESC LIMIT #{offset}, #{size}" +
            "</script>")
    List<Video> selectVideosByKeyword(Integer offset, Integer size, String keyword, Integer status);
    
    @Select("<script>" +
            "SELECT COUNT(*) FROM videos " +
            "<where>" +
            "  <if test='keyword != null and keyword != \"\"'>" +
            "    AND (title LIKE CONCAT('%', #{keyword}, '%') OR description LIKE CONCAT('%', #{keyword}, '%'))" +
            "  </if>" +
            "  <if test='status != null'>" +
            "    AND status = #{status}" +
            "  </if>" +
            "</where>" +
            "</script>")
    int countVideosByKeyword(String keyword, Integer status);
    
    @Select("SELECT * FROM videos WHERE id = #{id}")
    Video selectById(Integer id);
    
    @Update("UPDATE videos SET status = #{status} WHERE id = #{id}")
    int updateStatus(Integer id, Integer status);
    
    @Update("DELETE FROM videos WHERE id = #{id}")
    int delete(Integer id);
    
    @Select("SELECT * FROM videos ORDER BY id DESC")
    List<Video> selectAll();
    
    @Select("SELECT * FROM videos WHERE status = #{status} ORDER BY id DESC")
    List<Video> selectByStatus(Integer status);
    
    @Select("SELECT * FROM videos WHERE manuscript_id = #{manuscriptId} ORDER BY video_order ASC")
    List<Video> selectByManuscriptId(Integer manuscriptId);

    /**
     * 查询最近上架的视频（用于增量索引）
     * @param minutes 最近多少分钟内
     * @return 视频列表
     */
    @Select("SELECT * FROM videos WHERE status = 3 AND upload_time >= DATE_SUB(NOW(), INTERVAL #{minutes} MINUTE) ORDER BY id DESC")
    List<Video> selectRecentlyPublished(int minutes);
    
    @Update("<script>" +
            "UPDATE videos " +
            "<set>" +
            "  <if test='title != null'>title = #{title},</if>" +
            "  <if test='description != null'>description = #{description},</if>" +
            "  <if test='coverUrl != null'>cover_url = #{coverUrl},</if>" +
            "  <if test='playUrlHd != null'>play_url_hd = #{playUrlHd},</if>" +
            "  <if test='playUrlSd != null'>play_url_sd = #{playUrlSd},</if>" +
            "  <if test='playUrlLd != null'>play_url_ld = #{playUrlLd},</if>" +
            "  <if test='categoryId != null'>category_id = #{categoryId},</if>" +
            "  <if test='durationSeconds != null'>duration_seconds = #{durationSeconds},</if>" +
            "  <if test='status != null'>status = #{status},</if>" +
            "  <if test='reviewStatus != null'>review_status = #{reviewStatus},</if>" +
            "  <if test='reviewReason != null'>review_reason = #{reviewReason},</if>" +
            "  <if test='reviewTime != null'>review_time = #{reviewTime},</if>" +
            "  <if test='reviewerId != null'>reviewer_id = #{reviewerId},</if>" +
            "  <if test='processProgress != null'>process_progress = #{processProgress},</if>" +
            "  <if test='processStage != null'>process_stage = #{processStage},</if>" +
            "  <if test='processStatus != null'>process_status = #{processStatus},</if>" +
            "  <if test='processError != null'>process_error = #{processError},</if>" +
            "  <if test='sourceVideoUrl != null'>source_video_url = #{sourceVideoUrl},</if>" +
            "  <if test='hasSubtitle != null'>has_subtitle = #{hasSubtitle},</if>" +
            "  <if test='hasSummary != null'>has_summary = #{hasSummary},</if>" +
            "  updated_at = CURRENT_TIMESTAMP" +
            "</set>" +
            "WHERE id = #{id}" +
            "</script>")
    int update(Video video);
    
    // 统计相关方法
    @Select("SELECT DATE(create_time) as date, SUM(play_count) as playCount FROM videos WHERE create_time BETWEEN #{startDate} AND #{endDate} GROUP BY DATE(create_time) ORDER BY date")
    List<Map<String, Object>> getDailyPlayCount(String startDate, String endDate);
    
    @Select("SELECT SUM(play_count) FROM videos WHERE create_time BETWEEN #{startDate} AND #{endDate}")
    Integer getTotalPlayCount(String startDate, String endDate);
    
    @Select("SELECT id, title, play_count as playCount, comment_count as commentCount FROM videos ORDER BY play_count DESC LIMIT #{limit}")
    List<Map<String, Object>> getHotVideos(int limit);

    /**
     * 查询视频的标签列表
     * @param videoId 视频ID
     * @return 标签名称列表
     */
    @Select("SELECT t.name FROM tags t " +
            "INNER JOIN video_tags vt ON t.id = vt.tag_id " +
            "WHERE vt.video_id = #{videoId}")
    List<String> selectTagsByVideoId(Integer videoId);
}
