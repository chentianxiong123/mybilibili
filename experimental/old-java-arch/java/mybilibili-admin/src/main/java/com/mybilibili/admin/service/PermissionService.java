package com.mybilibili.admin.service;

import com.mybilibili.common.entity.Permission;

import java.util.List;

public interface PermissionService {
    List<Permission> getPermissionList();
    Permission getPermissionById(Integer id);
    boolean addPermission(Permission permission);
    boolean updatePermission(Permission permission);
    boolean deletePermission(Integer id);
}