package com.mybilibili.admin.mapper;

import com.mybilibili.common.entity.Manuscript;
import org.apache.ibatis.annotations.*;

import java.util.List;
import java.util.Map;

@Mapper
public interface ManuscriptMapper {

    @Select("SELECT * FROM manuscripts WHERE id = #{id}")
    Manuscript selectById(Integer id);

    @Select("SELECT * FROM manuscripts WHERE status = #{status} ORDER BY upload_time DESC")
    List<Manuscript> selectByStatus(Integer status);

    @Select("SELECT * FROM manuscripts ORDER BY upload_time DESC")
    List<Manuscript> selectAll();

    @Update("<script>" +
            "UPDATE manuscripts SET " +
            "title = #{title}, " +
            "description = #{description}, " +
            "cover_url = #{coverUrl}, " +
            "category_id = #{categoryId}, " +
            "status = #{status}, " +
            "review_status = #{reviewStatus}, " +
            "review_reason = #{reviewReason}, " +
            "review_time = #{reviewTime}, " +
            "reviewer_id = #{reviewerId}, " +
            "process_progress = #{processProgress}, " +
            "process_stage = #{processStage}, " +
            "duration_seconds = #{durationSeconds}, " +
            "updated_at = CURRENT_TIMESTAMP " +
            "WHERE id = #{id}" +
            "</script>")
    int update(Manuscript manuscript);

    @Update("UPDATE manuscripts SET status = #{status} WHERE id = #{id}")
    int updateStatus(@Param("id") Integer id, @Param("status") Integer status);

    @Select("SELECT COUNT(*) FROM manuscripts")
    int countAll();

    @Select("SELECT COUNT(*) FROM manuscripts WHERE status = #{status}")
    int countByStatus(Integer status);

    @Select("SELECT SUM(view_count) FROM manuscripts")
    Long sumTotalViewCount();

    @Select("SELECT status, COUNT(*) as count FROM manuscripts GROUP BY status")
    List<Map<String, Object>> countGroupByStatus();

    @Select("SELECT * FROM manuscripts ORDER BY upload_time DESC LIMIT #{limit}")
    List<Manuscript> selectRecent(@Param("limit") int limit);

    /**
     * 查询最近上架的稿件（用于增量索引）
     * @param minutes 最近多少分钟内
     * @return 稿件列表
     */
    @Select("SELECT * FROM manuscripts WHERE status = 3 AND upload_time >= DATE_SUB(NOW(), INTERVAL #{minutes} MINUTE) ORDER BY id DESC")
    List<Manuscript> selectRecentlyPublished(int minutes);
}
