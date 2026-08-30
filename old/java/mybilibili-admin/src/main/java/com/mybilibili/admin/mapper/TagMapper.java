package com.mybilibili.admin.mapper;

import com.mybilibili.common.entity.Tag;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Select;
import org.apache.ibatis.annotations.Update;

import java.util.List;

@Mapper
public interface TagMapper {
    @Select("SELECT * FROM tags WHERE name LIKE CONCAT('%', #{keyword}, '%') LIMIT #{offset}, #{size}")
    List<Tag> selectTagsByKeyword(Integer offset, Integer size, String keyword);
    
    @Select("SELECT COUNT(*) FROM tags WHERE name LIKE CONCAT('%', #{keyword}, '%')")
    int countTagsByKeyword(String keyword);
    
    @Select("SELECT * FROM tags WHERE id = #{id}")
    Tag selectById(Integer id);
    
    @Update("INSERT INTO tags (name, description, created_at, updated_at) VALUES (#{name}, #{description}, NOW(), NOW())")
    int insert(String name, String description);
    
    @Update("UPDATE tags SET name = #{name}, description = #{description}, updated_at = NOW() WHERE id = #{id}")
    int update(Integer id, String name, String description);
    
    @Update("DELETE FROM tags WHERE id = #{id}")
    int delete(Integer id);
}