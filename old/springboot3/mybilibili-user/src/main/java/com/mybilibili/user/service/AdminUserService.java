package com.mybilibili.user.service;

import com.mybilibili.common.dto.AdminLoginDTO;
import com.mybilibili.common.entity.AdminUser;
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
    Result<?> updateAdminUser(Integer id, String username, Integer adminLevel, String newPassword);
}