package com.mybilibili.admin.mapper;

import com.mybilibili.common.entity.Category;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Select;
import org.apache.ibatis.annotations.Update;

import java.util.List;

@Mapper
public interface CategoryMapper {
    @Select("SELECT * FROM categories WHERE name LIKE CONCAT('%', #{keyword}, '%') LIMIT #{offset}, #{size}")
    List<Category> selectCategoriesByKeyword(Integer offset, Integer size, String keyword);
    
    @Select("SELECT COUNT(*) FROM categories WHERE name LIKE CONCAT('%', #{keyword}, '%')")
    int countCategoriesByKeyword(String keyword);
    
    @Select("SELECT * FROM categories WHERE id = #{id}")
    Category selectById(Integer id);
    
    @Update("INSERT INTO categories (name, created_at, updated_at) VALUES (#{name}, NOW(), NOW())")
    int insert(String name);

    @Update("UPDATE categories SET name = #{name}, updated_at = NOW() WHERE id = #{id}")
    int update(Integer id, String name);
    
    @Update("DELETE FROM categories WHERE id = #{id}")
    int delete(Integer id);
}