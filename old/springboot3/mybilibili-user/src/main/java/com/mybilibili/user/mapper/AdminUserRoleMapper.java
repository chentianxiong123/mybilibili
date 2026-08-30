package com.mybilibili.user.mapper;

import com.mybilibili.common.entity.AdminUserRole;
import com.mybilibili.common.entity.Role;
import org.apache.ibatis.annotations.*;

import java.util.List;

@Mapper
public interface AdminUserRoleMapper {

    @Insert("INSERT INTO admin_user_roles (admin_user_id, role_id) VALUES (#{adminUserId}, #{roleId})")
    int insert(AdminUserRole adminUserRole);

    @Delete("DELETE FROM admin_user_roles WHERE admin_user_id = #{adminUserId} AND role_id = #{roleId}")
    int delete(AdminUserRole adminUserRole);

    @Delete("DELETE FROM admin_user_roles WHERE admin_user_id = #{adminUserId}")
    int deleteByAdminUserId(Integer adminUserId);

    @Select("SELECT * FROM admin_user_roles WHERE admin_user_id = #{adminUserId}")
    List<AdminUserRole> selectByAdminUserId(Integer adminUserId);

    @Select("SELECT r.* FROM roles r JOIN admin_user_roles aur ON r.id = aur.role_id WHERE aur.admin_user_id = #{adminUserId}")
    List<Role> selectRolesByAdminUserId(Integer adminUserId);
}