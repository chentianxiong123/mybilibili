package com.mybilibili.search.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.common.entity.Manuscript;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

import java.util.List;

public interface ManuscriptMapper extends BaseMapper<Manuscript> {

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