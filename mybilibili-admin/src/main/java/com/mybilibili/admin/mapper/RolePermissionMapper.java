package com.mybilibili.admin.mapper;

import com.mybilibili.common.entity.RolePermission;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Insert;
import org.apache.ibatis.annotations.Delete;
import org.apache.ibatis.annotations.Select;

import java.util.List;

@Mapper
public interface RolePermissionMapper {
    @Insert("INSERT INTO role_permissions (role_id, permission_id) VALUES (#{roleId}, #{permissionId})")
    int insert(RolePermission rolePermission);
    
    @Delete("DELETE FROM role_permissions WHERE role_id = #{roleId} AND permission_id = #{permissionId}")
    int delete(RolePermission rolePermission);
    
    @Delete("DELETE FROM role_permissions WHERE role_id = #{roleId}")
    int deleteByRoleId(Integer roleId);
    
    @Select("SELECT * FROM role_permissions WHERE role_id = #{roleId}")
    List<RolePermission> selectByRoleId(Integer roleId);
}