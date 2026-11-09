package com.mybilibili.admin.service;

import com.mybilibili.admin.dto.AdminLoginDTO;
import com.mybilibili.admin.entity.AdminUser;
import com.mybilibili.common.entity.Role;
import com.mybilibili.common.vo.Result;

import java.util.List;
import java.util.Map;

public interface AdminUserService {
    Result<?> login(AdminLoginDTO adminLoginDTO);
    Result<?> register(AdminLoginDTO adminLoginDTO);
    List<Map<String, Object>> getAdminUserList();
    AdminUser getAdminUserById(Integer id);
    boolean updateAdminUserRoles(Integer adminUserId, List<Integer> roleIds);
    List<Role> getAdminUserRoles(Integer adminUserId);
}