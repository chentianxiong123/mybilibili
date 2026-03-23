package com.mybilibili.admin.mapper;

import com.mybilibili.common.entity.Permission;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Select;
import org.apache.ibatis.annotations.Insert;
import org.apache.ibatis.annotations.Update;
import org.apache.ibatis.annotations.Delete;

import java.util.List;

@Mapper
public interface PermissionMapper {
    @Select("SELECT * FROM permissions")
    List<Permission> selectAll();
    
    @Select("SELECT * FROM permissions WHERE id = #{id}")
    Permission selectById(Integer id);
    
    @Insert("INSERT INTO permissions (name, code, url, method, parent_id, description, create_time, update_time) VALUES (#{name}, #{code}, #{url}, #{method}, #{parentId}, #{description}, NOW(), NOW())")
    int insert(Permission permission);
    
    @Update("UPDATE permissions SET name = #{name}, code = #{code}, url = #{url}, method = #{method}, parent_id = #{parentId}, description = #{description}, update_time = NOW() WHERE id = #{id}")
    int update(Permission permission);
    
    @Delete("DELETE FROM permissions WHERE id = #{id}")
    int delete(Integer id);
    
    @Select("SELECT p.* FROM permissions p JOIN role_permissions rp ON p.id = rp.permission_id WHERE rp.role_id = #{roleId}")
    List<Permission> selectByRoleId(Integer roleId);
}