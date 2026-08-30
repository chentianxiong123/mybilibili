package com.mybilibili.user.mapper;

import com.mybilibili.common.entity.Permission;
import org.apache.ibatis.annotations.*;

import java.util.List;

@Mapper
public interface RolePermissionMapper {

    @Insert("INSERT INTO role_permissions (role_id, permission_id) VALUES (#{roleId}, #{permissionId})")
    int insert(@Param("roleId") Integer roleId, @Param("permissionId") Integer permissionId);

    @Delete("DELETE FROM role_permissions WHERE role_id = #{roleId}")
    int deleteByRoleId(Integer roleId);

    @Select("SELECT p.* FROM permissions p JOIN role_permissions rp ON p.id = rp.permission_id WHERE rp.role_id = #{roleId}")
    List<Permission> selectPermissionsByRoleId(Integer roleId);
}
