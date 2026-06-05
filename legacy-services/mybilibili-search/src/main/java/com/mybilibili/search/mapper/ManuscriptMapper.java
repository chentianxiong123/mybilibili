package com.mybilibili.search.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.common.entity.Manuscript;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

import java.util.List;
import java.util.Map;

public interface ManuscriptMapper extends BaseMapper<Manuscript> {

    @Select("SELECT m.*, " +
            "u.nickname as userName, " +
            "c.name as categoryName, " +
            "GROUP_CONCAT(DISTINCT v.title SEPARATOR '||') as videoTitles, " +
            "GROUP_CONCAT(DISTINCT v.id SEPARATOR ',') as videoIds, " +
            "COUNT(DISTINCT v.id) as videoCount " +
            "FROM manuscripts m " +
            "LEFT JOIN users u ON m.user_id = u.id " +
            "LEFT JOIN categories c ON m.category_id = c.id " +
            "LEFT JOIN videos v ON v.manuscript_id = m.id " +
            "WHERE m.status = 3 " +
            "GROUP BY m.id")
    List<Map<String, Object>> selectPublishedManuscripts();

    @Select("SELECT DISTINCT t.name FROM tags t " +
            "INNER JOIN video_tags vt ON t.id = vt.tag_id " +
            "INNER JOIN videos v ON vt.video_id = v.id " +
            "WHERE v.manuscript_id = #{manuscriptId}")
    List<String> selectTagsByManuscriptId(@Param("manuscriptId") Integer manuscriptId);

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

    @Select("SELECT * FROM manuscripts WHERE status = #{status} AND title LIKE CONCAT('%', #{keyword}, '%') " +
            "ORDER BY view_count DESC LIMIT #{limit}")
    List<Manuscript> suggestByKeyword(@Param("keyword") String keyword,
                                      @Param("status") Integer status,
                                      @Param("limit") Integer limit);
}