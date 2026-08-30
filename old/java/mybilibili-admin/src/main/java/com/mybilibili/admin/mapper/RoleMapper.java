package com.mybilibili.admin.mapper;

import com.mybilibili.common.entity.Role;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Select;
import org.apache.ibatis.annotations.Insert;
import org.apache.ibatis.annotations.Update;
import org.apache.ibatis.annotations.Delete;

import java.util.List;

@Mapper
public interface RoleMapper {
    @Select("SELECT * FROM roles")
    List<Role> selectAll();
    
    @Select("SELECT * FROM roles WHERE id = #{id}")
    Role selectById(Integer id);
    
    @Insert("INSERT INTO roles (name, description, create_time, update_time) VALUES (#{name}, #{description}, NOW(), NOW())")
    int insert(Role role);
    
    @Update("UPDATE roles SET name = #{name}, description = #{description}, update_time = NOW() WHERE id = #{id}")
    int update(Role role);
    
    @Delete("DELETE FROM roles WHERE id = #{id}")
    int delete(Integer id);
}