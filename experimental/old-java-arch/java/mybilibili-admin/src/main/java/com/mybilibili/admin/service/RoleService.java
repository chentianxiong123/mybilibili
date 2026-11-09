package com.mybilibili.admin.service;

import com.mybilibili.common.entity.Role;
import com.mybilibili.common.entity.Permission;

import java.util.List;

public interface RoleService {
    List<Role> getRoleList();
    Role getRoleById(Integer id);
    boolean addRole(Role role);
    boolean updateRole(Role role);
    boolean deleteRole(Integer id);
    List<Permission> getRolePermissions(Integer roleId);
    boolean setRolePermissions(Integer roleId, List<Integer> permissionIds);
}