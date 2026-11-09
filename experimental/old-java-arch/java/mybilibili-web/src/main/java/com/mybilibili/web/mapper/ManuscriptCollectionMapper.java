package com.mybilibili.web.mapper;

import com.mybilibili.common.entity.ManuscriptCollection;
import org.apache.ibatis.annotations.*;

import java.util.List;

@Mapper
public interface ManuscriptCollectionMapper {

    @Insert("INSERT INTO manuscript_collections (title, description, cover_url, user_id, manuscript_count, view_count, status, created_at, updated_at) " +
            "VALUES (#{title}, #{description}, #{coverUrl}, #{userId}, #{manuscriptCount}, #{viewCount}, #{status}, #{createdAt}, #{updatedAt})")
    @Options(useGeneratedKeys = true, keyProperty = "id")
    int insert(ManuscriptCollection collection);

    @Select("SELECT * FROM manuscript_collections WHERE id = #{id}")
    ManuscriptCollection selectById(Integer id);

    @Select("SELECT * FROM manuscript_collections WHERE user_id = #{userId} ORDER BY created_at DESC")
    List<ManuscriptCollection> selectByUserId(Integer userId);

    @Select("SELECT * FROM manuscript_collections WHERE user_id = #{userId} AND status = #{status} ORDER BY created_at DESC")
    List<ManuscriptCollection> selectByUserIdAndStatus(@Param("userId") Integer userId, @Param("status") Integer status);

    @Update("UPDATE manuscript_collections SET " +
            "title = #{title}, " +
            "description = #{description}, " +
            "cover_url = #{coverUrl}, " +
            "status = #{status}, " +
            "updated_at = CURRENT_TIMESTAMP " +
            "WHERE id = #{id}")
    int update(ManuscriptCollection collection);

    @Update("UPDATE manuscript_collections SET " +
            "manuscript_count = manuscript_count + #{count}, " +
            "updated_at = CURRENT_TIMESTAMP " +
            "WHERE id = #{id}")
    int updateManuscriptCount(@Param("id") Integer id, @Param("count") Integer count);

    @Update("UPDATE manuscript_collections SET " +
            "view_count = view_count + #{count}, " +
            "updated_at = CURRENT_TIMESTAMP " +
            "WHERE id = #{id}")
    int updateViewCount(@Param("id") Integer id, @Param("count") Integer count);

    @Delete("DELETE FROM manuscript_collections WHERE id = #{id}")
    int delete(Integer id);

    /**
     * 检查用户是否有权限操作合集
     */
    @Select("SELECT COUNT(*) FROM manuscript_collections WHERE id = #{collectionId} AND user_id = #{userId}")
    int checkOwnership(@Param("collectionId") Integer collectionId, @Param("userId") Integer userId);
}
