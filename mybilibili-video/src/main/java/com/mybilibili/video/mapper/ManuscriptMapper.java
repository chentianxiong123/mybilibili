package com.mybilibili.video.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.common.entity.Manuscript;
import org.apache.ibatis.annotations.*;

import java.util.List;

@Mapper
public interface ManuscriptMapper extends BaseMapper<Manuscript> {

    @Select("SELECT * FROM manuscripts WHERE id = #{id}")
    Manuscript selectById(Integer id);

    @Select("SELECT * FROM manuscripts WHERE user_id = #{userId} ORDER BY upload_time DESC")
    List<Manuscript> selectByUserId(Integer userId);

    @Select("<script>" +
            "SELECT * FROM manuscripts WHERE user_id = #{userId} " +
            "<if test='status != null'> AND status = #{status} </if>" +
            "ORDER BY upload_time DESC LIMIT #{offset}, #{size}" +
            "</script>")
    List<Manuscript> selectByUserIdWithPaging(@Param("userId") Integer userId, @Param("status") Integer status, @Param("offset") Integer offset, @Param("size") Integer size);

    @Select("<script>" +
            "SELECT COUNT(*) FROM manuscripts WHERE user_id = #{userId} " +
            "<if test='status != null'> AND status = #{status} </if>" +
            "</script>")
    Integer countByUserIdAndStatus(@Param("userId") Integer userId, @Param("status") Integer status);

    @Select("SELECT * FROM manuscripts WHERE status = #{status} ORDER BY upload_time DESC")
    List<Manuscript> selectByStatus(@Param("status") Integer status);

    @Select("SELECT * FROM manuscripts WHERE status = 3 ORDER BY upload_time DESC LIMIT #{offset}, #{size}")
    List<Manuscript> selectRecommended(@Param("offset") Integer offset, @Param("size") Integer size);

    @Insert("INSERT INTO manuscripts (title, description, cover_url, user_id, category_id, view_count, like_count, coin_count, collect_count, share_count, comment_count, danmaku_count, duration, duration_seconds, status, review_status, upload_time) " +
            "VALUES (#{title}, #{description}, #{coverUrl}, #{userId}, #{categoryId}, 0, 0, 0, 0, 0, 0, 0, #{duration}, #{durationSeconds}, #{status}, #{reviewStatus}, #{uploadTime})")
    @Options(useGeneratedKeys = true, keyProperty = "id")
    int insert(Manuscript manuscript);

    @Update("UPDATE manuscripts SET title = #{title}, description = #{description}, cover_url = #{coverUrl}, category_id = #{categoryId}, duration = #{duration}, duration_seconds = #{durationSeconds}, status = #{status}, updated_at = NOW() WHERE id = #{id}")
    int updateById(Manuscript manuscript);

    @Delete("DELETE FROM manuscripts WHERE id = #{id}")
    int deleteById(Integer id);

    @Select("SELECT COUNT(*) FROM manuscripts WHERE user_id = #{userId} AND status = 0")
    Integer countPendingByUserId(Integer userId);

    @Select("SELECT COUNT(*) FROM manuscripts WHERE user_id = #{userId} AND status = 1")
    Integer countProcessingByUserId(Integer userId);

    @Select("SELECT COUNT(*) FROM manuscripts WHERE user_id = #{userId} AND status = 2")
    Integer countReadyByUserId(Integer userId);

    @Select("SELECT COUNT(*) FROM manuscripts WHERE user_id = #{userId} AND status = 3")
    Integer countPublishedByUserId(Integer userId);

    @Select("SELECT COUNT(*) FROM manuscripts WHERE user_id = #{userId} AND status = 4")
    Integer countRejectedByUserId(Integer userId);

    @Select("SELECT * FROM manuscripts WHERE category_id = #{categoryId} AND status = 3 ORDER BY upload_time DESC LIMIT #{offset}, #{size}")
    List<Manuscript> selectByCategoryId(@Param("categoryId") Integer categoryId, @Param("offset") Integer offset, @Param("size") Integer size);

    @Select("SELECT COUNT(*) FROM manuscripts WHERE category_id = #{categoryId} AND status = 3")
    Integer countByCategoryId(Integer categoryId);

    @Select("SELECT * FROM manuscripts WHERE status = 3 ORDER BY view_count DESC LIMIT #{offset}, #{size}")
    List<Manuscript> selectHot(@Param("offset") Integer offset, @Param("size") Integer size);

    @Select("SELECT id FROM manuscripts WHERE user_id = #{userId}")
    List<Integer> selectIdsByUserId(@Param("userId") Integer userId);

    @Select("SELECT COUNT(*) FROM manuscripts")
    Integer countAll();

    @Select("SELECT COALESCE(SUM(view_count), 0) FROM manuscripts")
    Integer selectSumViewCount();

    @Update("UPDATE manuscripts SET share_count = share_count + #{count} WHERE id = #{manuscriptId}")
    int updateShareCount(@Param("manuscriptId") Integer manuscriptId, @Param("count") Integer count);

    @Update("UPDATE manuscripts SET view_count = view_count + 1 WHERE id = #{id}")
    int incrementViewCount(@Param("id") Integer id);

    @Update("UPDATE manuscripts SET status = #{status}, updated_at = NOW() WHERE id = #{id}")
    int updateStatusById(@Param("id") Integer id, @Param("status") Integer status);

    @Select("SELECT * FROM manuscripts WHERE user_id = #{userId} AND title LIKE #{keyword} AND status = 3")
    List<Manuscript> searchByUserIdAndKeyword(@Param("userId") Integer userId, @Param("keyword") String keyword);

    @Update("UPDATE manuscripts SET comment_count = #{count}, updated_at = NOW() WHERE id = #{id}")
    int updateCommentCount(@Param("id") Integer id, @Param("count") Integer count);

    @Update("UPDATE manuscripts SET comment_count = comment_count + 1, updated_at = NOW() WHERE id = #{id}")
    int incrementCommentCount(@Param("id") Integer id);

    @Update("UPDATE manuscripts SET comment_count = GREATEST(0, comment_count - 1), updated_at = NOW() WHERE id = #{id}")
    int decrementCommentCount(@Param("id") Integer id);
}
