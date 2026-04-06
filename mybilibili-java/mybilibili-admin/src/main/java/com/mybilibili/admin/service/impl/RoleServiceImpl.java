package com.mybilibili.admin.service.impl;

import com.mybilibili.admin.service.RoleService;
import com.mybilibili.admin.mapper.RoleMapper;
import com.mybilibili.admin.mapper.PermissionMapper;
import com.mybilibili.admin.mapper.RolePermissionMapper;
import com.mybilibili.common.entity.Role;
import com.mybilibili.common.entity.Permission;
import com.mybilibili.common.entity.RolePermission;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class RoleServiceImpl implements RoleService {

    @Autowired
    private RoleMapper roleMapper;

    @Autowired
    private PermissionMapper permissionMapper;

    @Autowired
    private RolePermissionMapper rolePermissionMapper;

    @Override
    public List<Role> getRoleList() {
        return roleMapper.selectAll();
    }

    @Override
    public Role getRoleById(Integer id) {
        return roleMapper.selectById(id);
    }

    @Override
    public boolean addRole(Role role) {
        return roleMapper.insert(role) > 0;
    }

    @Override
    public boolean updateRole(Role role) {
        return roleMapper.update(role) > 0;
    }

    @Override
    public boolean deleteRole(Integer id) {
        // 先删除角色与权限的关联关系
        rolePermissionMapper.deleteByRoleId(id);
        // 再删除角色
        return roleMapper.delete(id) > 0;
    }

    @Override
    public List<Permission> getRolePermissions(Integer roleId) {
        return permissionMapper.selectByRoleId(roleId);
    }

    @Override
    public boolean setRolePermissions(Integer roleId, List<Integer> permissionIds) {
        // 先删除原有的权限关联
        rolePermissionMapper.deleteByRoleId(roleId);
        // 添加新的权限关联
        for (Integer permissionId : permissionIds) {
            RolePermission rolePermission = new RolePermission();
            rolePermission.setRoleId(roleId);
            rolePermission.setPermissionId(permissionId);
            rolePermissionMapper.insert(rolePermission);
        }
        return true;
    }
}