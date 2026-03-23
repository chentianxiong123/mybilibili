package com.mybilibili.admin.service.impl;

import com.mybilibili.admin.service.PermissionService;
import com.mybilibili.admin.mapper.PermissionMapper;
import com.mybilibili.common.entity.Permission;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class PermissionServiceImpl implements PermissionService {

    @Autowired
    private PermissionMapper permissionMapper;

    @Override
    public List<Permission> getPermissionList() {
        return permissionMapper.selectAll();
    }

    @Override
    public Permission getPermissionById(Integer id) {
        return permissionMapper.selectById(id);
    }

    @Override
    public boolean addPermission(Permission permission) {
        return permissionMapper.insert(permission) > 0;
    }

    @Override
    public boolean updatePermission(Permission permission) {
        return permissionMapper.update(permission) > 0;
    }

    @Override
    public boolean deletePermission(Integer id) {
        return permissionMapper.delete(id) > 0;
    }
}