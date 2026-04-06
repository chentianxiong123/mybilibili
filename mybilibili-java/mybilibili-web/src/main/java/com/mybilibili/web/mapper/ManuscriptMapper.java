package com.mybilibili.web.mapper;

import com.mybilibili.common.entity.Manuscript;
import org.apache.ibatis.annotations.*;

import java.util.List;
import java.util.Map;

@Mapper
public interface ManuscriptMapper {

    @Insert("INSERT INTO manuscripts (title, description, cover_url, user_id, category_id, " +
            "status, review_status, process_progress, upload_time) " +
            "VALUES (#{title}, #{description}, #{coverUrl}, #{userId}, #{categoryId}, " +
            "#{status}, #{reviewStatus}, #{processProgress}, #{uploadTime})")
    @Options(useGeneratedKeys = true, keyProperty = "id")
    int insert(Manuscript manuscript);

    @Select("SELECT * FROM manuscripts WHERE id = #{id}")
    Manuscript selectById(Integer id);

    @Select("SELECT * FROM manuscripts WHERE user_id = #{userId} ORDER BY upload_time DESC")
    List<Manuscript> selectByUserId(Integer userId);

    @Select("SELECT * FROM manuscripts WHERE status = #{status} ORDER BY upload_time DESC")
    List<Manuscript> selectByStatus(Integer status);



    @Select("SELECT * FROM manuscripts WHERE category_id = #{categoryId} AND status = 3 ORDER BY upload_time DESC")
    List<Manuscript> selectByCategoryId(Integer categoryId);

    @Select("SELECT * FROM manuscripts ORDER BY upload_time DESC")
    List<Manuscript> selectAll();

    @Update("UPDATE manuscripts SET " +
            "title = #{title}, " +
            "description = #{description}, " +
            "cover_url = #{coverUrl}, " +
            "category_id = #{categoryId}, " +
            "status = #{status}, " +
            "review_status = #{reviewStatus}, " +
            "review_reason = #{reviewReason}, " +
            "review_time = #{reviewTime}, " +
            "reviewer_id = #{reviewerId}, " +
            "updated_at = CURRENT_TIMESTAMP " +
            "WHERE id = #{id}")
    int update(Manuscript manuscript);

    @Delete("DELETE FROM manuscripts WHERE id = #{id}")
    int delete(Integer id);

    @Update("UPDATE manuscripts SET view_count = GREATEST(0, view_count + #{count}) WHERE id = #{id}")
    int updateViewCount(@Param("id") Integer id, @Param("count") Integer count);

    @Update("UPDATE manuscripts SET like_count = GREATEST(0, like_count + #{count}) WHERE id = #{id}")
    int updateLikeCount(@Param("id") Integer id, @Param("count") Integer count);

    @Update("UPDATE manuscripts SET coin_count = GREATEST(0, coin_count + #{count}) WHERE id = #{id}")
    int updateCoinCount(@Param("id") Integer id, @Param("count") Integer count);

    @Update("UPDATE manuscripts SET collect_count = GREATEST(0, collect_count + #{count}) WHERE id = #{id}")
    int updateCollectCount(@Param("id") Integer id, @Param("count") Integer count);

    @Update("UPDATE manuscripts SET share_count = GREATEST(0, share_count + #{count}) WHERE id = #{id}")
    int updateShareCount(@Param("id") Integer id, @Param("count") Integer count);

    @Update("UPDATE manuscripts SET danmaku_count = GREATEST(0, danmaku_count + #{count}) WHERE id = #{id}")
    int updateDanmakuCount(@Param("id") Integer id, @Param("count") Integer count);

    /**
     * 统计用户所有稿件的总播放数
     * @param userId 用户ID
     * @return 总播放数
     */
    @Select("SELECT COALESCE(SUM(view_count), 0) FROM manuscripts WHERE user_id = #{userId}")
    Integer sumViewCountByUserId(Integer userId);

    /**
     * 统计用户所有稿件的总获赞数
     * @param userId 用户ID
     * @return 总获赞数
     */
    @Select("SELECT COALESCE(SUM(like_count), 0) FROM manuscripts WHERE user_id = #{userId}")
    Integer sumLikeCountByUserId(Integer userId);

    /**
     * 分页查询用户稿件列表（支持状态筛选）
     * @param userId 用户ID
     * @param status 状态（可选）
     * @param offset 偏移量
     * @param size 每页数量
     * @return 稿件列表
     */
    @Select("<script>" +
            "SELECT * FROM manuscripts WHERE user_id = #{userId} " +
            "<if test='status != null'> AND status = #{status} </if>" +
            "ORDER BY upload_time DESC " +
            "LIMIT #{offset}, #{size}" +
            "</script>")
    List<Manuscript> selectByUserIdWithPaging(@Param("userId") Integer userId,
                                               @Param("status") Integer status,
                                               @Param("offset") Integer offset,
                                               @Param("size") Integer size);

    /**
     * 统计用户稿件数量（支持状态筛选）
     * @param userId 用户ID
     * @param status 状态（可选）
     * @return 稿件数量
     */
    @Select("<script>" +
            "SELECT COUNT(*) FROM manuscripts WHERE user_id = #{userId} " +
            "<if test='status != null'> AND status = #{status} </if>" +
            "</script>")
    Integer countByUserIdAndStatus(@Param("userId") Integer userId, @Param("status") Integer status);

    /**
     * 统计用户各状态稿件数量
     * @param userId 用户ID
     * @return 各状态稿件数量统计
     */
    @Select("SELECT status, COUNT(*) as count FROM manuscripts WHERE user_id = #{userId} GROUP BY status")
    List<Map<String, Object>> countByUserIdGroupByStatus(@Param("userId") Integer userId);

    /**
     * 更新稿件时长
     * @param id 稿件ID
     * @param durationSeconds 时长（秒）
     * @return 影响行数
     */
    @Update("UPDATE manuscripts SET duration_seconds = #{durationSeconds}, updated_at = CURRENT_TIMESTAMP WHERE id = #{id}")
    int updateDuration(@Param("id") Integer id, @Param("durationSeconds") Integer durationSeconds);

    /**
     * MySQL搜索 - 根据关键词搜索稿件
     * @param keyword 关键词
     * @param categoryId 分类ID
     * @param userId 用户ID
     * @param status 状态
     * @param orderBy 排序
     * @param offset 偏移量
     * @param size 数量
     * @return 稿件列表
     */
    @Select("<script>" +
            "SELECT * FROM manuscripts WHERE status = #{status} " +
            "<if test='keyword != null and keyword != \"\"'> AND (title LIKE CONCAT('%', #{keyword}, '%') OR description LIKE CONCAT('%', #{keyword}, '%')) </if>" +
            "<if test='categoryId != null'> AND category_id = #{categoryId} </if>" +
            "<if test='userId != null'> AND user_id = #{userId} </if>" +
            "ORDER BY ${orderBy} " +
            "LIMIT #{offset}, #{size}" +
            "</script>")
    List<Manuscript> searchByKeyword(@Param("keyword") String keyword,
                                      @Param("categoryId") Integer categoryId,
                                      @Param("userId") Integer userId,
                                      @Param("status") Integer status,
                                      @Param("orderBy") String orderBy,
                                      @Param("offset") Integer offset,
                                      @Param("size") Integer size);

    /**
     * MySQL搜索 - 统计搜索结果数量
     * @param keyword 关键词
     * @param categoryId 分类ID
     * @param userId 用户ID
     * @param status 状态
     * @return 数量
     */
    @Select("<script>" +
            "SELECT COUNT(*) FROM manuscripts WHERE status = #{status} " +
            "<if test='keyword != null and keyword != \"\"'> AND (title LIKE CONCAT('%', #{keyword}, '%') OR description LIKE CONCAT('%', #{keyword}, '%')) </if>" +
            "<if test='categoryId != null'> AND category_id = #{categoryId} </if>" +
            "<if test='userId != null'> AND user_id = #{userId} </if>" +
            "</script>")
    long countSearchByKeyword(@Param("keyword") String keyword,
                               @Param("categoryId") Integer categoryId,
                               @Param("userId") Integer userId,
                               @Param("status") Integer status);

    /**
     * MySQL搜索 - 搜索建议
     * @param keyword 关键词
     * @param status 状态
     * @param limit 限制数量
     * @return 稿件列表
     */
    @Select("SELECT * FROM manuscripts WHERE status = #{status} AND title LIKE CONCAT('%', #{keyword}, '%') " +
            "ORDER BY view_count DESC LIMIT #{limit}")
    List<Manuscript> suggestByKeyword(@Param("keyword") String keyword,
                                       @Param("status") Integer status,
                                       @Param("limit") Integer limit);

    /**
     * MySQL推荐 - 查询相关稿件（同分类或相似标题）
     * @param excludeId 排除的稿件ID
     * @param categoryId 分类ID
     * @param title 标题
     * @param status 状态
     * @param limit 限制数量
     * @return 稿件列表
     */
    @Select("SELECT * FROM manuscripts WHERE status = #{status} AND id != #{excludeId} " +
            "AND (category_id = #{categoryId} OR title LIKE CONCAT('%', #{title}, '%')) " +
            "ORDER BY view_count DESC LIMIT #{limit}")
    List<Manuscript> selectRelatedManuscripts(@Param("excludeId") Integer excludeId,
                                               @Param("categoryId") Integer categoryId,
                                               @Param("title") String title,
                                               @Param("status") Integer status,
                                               @Param("limit") Integer limit);

    /**
     * MySQL推荐 - 查询热门稿件
     * @param categoryId 分类ID
     * @param status 状态
     * @param limit 限制数量
     * @return 稿件列表
     */
    @Select("<script>" +
            "SELECT * FROM manuscripts WHERE status = #{status} " +
            "<if test='categoryId != null'> AND category_id = #{categoryId} </if> " +
            "ORDER BY view_count DESC LIMIT #{limit}" +
            "</script>")
    List<Manuscript> selectHotManuscripts(@Param("categoryId") Integer categoryId,
                                           @Param("status") Integer status,
                                           @Param("limit") Integer limit);

    /**
     * MySQL推荐 - 查询个性化推荐稿件
     * @param excludeIds 排除的稿件ID列表
     * @param categoryIds 感兴趣的分类ID列表
     * @param status 状态
     * @param limit 限制数量
     * @return 稿件列表
     */
    @Select("<script>" +
            "SELECT * FROM manuscripts WHERE status = #{status} " +
            "<if test='excludeIds != null and excludeIds.size() > 0'> " +
            "AND id NOT IN " +
            "<foreach collection='excludeIds' item='id' open='(' separator=',' close=')'> #{id} </foreach> " +
            "</if> " +
            "<if test='categoryIds != null and categoryIds.size() > 0'> " +
            "AND category_id IN " +
            "<foreach collection='categoryIds' item='id' open='(' separator=',' close=')'> #{id} </foreach> " +
            "</if> " +
            "ORDER BY view_count DESC LIMIT #{limit}" +
            "</script>")
    List<Manuscript> selectRecommendedManuscripts(@Param("excludeIds") List<Integer> excludeIds,
                                                   @Param("categoryIds") List<Integer> categoryIds,
                                                   @Param("status") Integer status,
                                                   @Param("limit") Integer limit);
}
