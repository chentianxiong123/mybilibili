package com.mybilibili.interaction.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.common.entity.ManuscriptCollection;
import org.apache.ibatis.annotations.Delete;
import org.apache.ibatis.annotations.Insert;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Options;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;
import org.apache.ibatis.annotations.Update;

import java.util.List;

@Mapper
public interface ManuscriptCollectionMapper extends BaseMapper<ManuscriptCollection> {

    @Select("SELECT * FROM manuscript_collections WHERE user_id = #{userId} ORDER BY created_at DESC")
    List<ManuscriptCollection> selectByUserId(@Param("userId") Integer userId);

    @Select("SELECT * FROM manuscript_collections WHERE id = #{id}")
    ManuscriptCollection selectById(@Param("id") Integer id);

    @Update("UPDATE manuscript_collections SET title=#{title}, description=#{description}, cover_url=#{coverUrl}, status=#{status}, updated_at=NOW() WHERE id=#{id}")
    int update(ManuscriptCollection collection);

    @Insert("INSERT INTO manuscript_collections (title, description, cover_url, user_id, manuscript_count, view_count, status, created_at, updated_at) VALUES (#{title}, #{description}, #{coverUrl}, #{userId}, #{manuscriptCount}, #{viewCount}, #{status}, NOW(), NOW())")
    @Options(useGeneratedKeys = true, keyProperty = "id")
    int insert(ManuscriptCollection collection);

    @Delete("DELETE FROM manuscript_collections WHERE id = #{id}")
    int deleteById(@Param("id") Integer id);
}
