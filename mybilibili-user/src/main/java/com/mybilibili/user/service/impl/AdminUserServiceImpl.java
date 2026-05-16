package com.mybilibili.user.service.impl;

import com.mybilibili.common.dto.AdminLoginDTO;
import com.mybilibili.common.entity.AdminUser;
import com.mybilibili.common.entity.AdminUserRole;
import com.mybilibili.common.entity.Role;
import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.vo.Result;
import com.mybilibili.user.mapper.AdminUserMapper;
import com.mybilibili.user.mapper.AdminUserRoleMapper;
import com.mybilibili.user.service.AdminUserService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

@Service
public class AdminUserServiceImpl implements AdminUserService {

    @Autowired
    private AdminUserMapper adminUserMapper;

    @Autowired
    private AdminUserRoleMapper adminUserRoleMapper;

    private final BCryptPasswordEncoder passwordEncoder = new BCryptPasswordEncoder();

    @Override
    public Result<?> login(AdminLoginDTO adminLoginDTO) {
        try {
            AdminUser adminUser = adminUserMapper.selectByUsername(adminLoginDTO.getUsername());
            if (adminUser == null) {
                return Result.error("管理员不存在");
            }

            if (!passwordEncoder.matches(adminLoginDTO.getPassword(), adminUser.getPassword())) {
                return Result.error("密码错误");
            }

            String token = JwtUtils.generateToken(adminUser.getId(), adminUser.getUsername());

            Map<String, Object> data = new HashMap<>();
            data.put("token", token);
            data.put("adminUser", adminUser);

            return Result.success("登录成功", data);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<?> register(AdminLoginDTO adminLoginDTO) {
        try {
            AdminUser existingAdmin = adminUserMapper.selectByUsername(adminLoginDTO.getUsername());
            if (existingAdmin != null) {
                return Result.error("用户名已存在");
            }

            String encryptedPassword = passwordEncoder.encode(adminLoginDTO.getPassword());

            AdminUser adminUser = new AdminUser();
            adminUser.setUsername(adminLoginDTO.getUsername());
            adminUser.setPassword(encryptedPassword);
            adminUser.setAdminLevel(1);

            int result = adminUserMapper.insert(adminUser);
            if (result > 0) {
                return Result.success("注册成功", null);
            } else {
                return Result.error("注册失败");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public List<Map<String, Object>> getAdminUserList() {
        List<AdminUser> adminUsers = adminUserMapper.selectAll();
        List<Map<String, Object>> result = new ArrayList<>();

        for (AdminUser adminUser : adminUsers) {
            Map<String, Object> adminMap = new HashMap<>();
            adminMap.put("id", adminUser.getId());
            adminMap.put("username", adminUser.getUsername());
            adminMap.put("adminLevel", adminUser.getAdminLevel());
            adminMap.put("createdAt", adminUser.getCreatedAt());
            adminMap.put("updatedAt", adminUser.getUpdatedAt());

            List<Role> roles = adminUserRoleMapper.selectRolesByAdminUserId(adminUser.getId());
            adminMap.put("roles", roles);

            String roleNames = roles.stream()
                    .map(Role::getName)
                    .collect(Collectors.joining(", "));
            adminMap.put("roleNames", roleNames.isEmpty() ? "无角色" : roleNames);

            result.add(adminMap);
        }

        return result;
    }

    @Override
    public AdminUser getAdminUserById(Integer id) {
        return adminUserMapper.selectById(id);
    }

    @Override
    public boolean updateAdminUserRoles(Integer adminUserId, List<Integer> roleIds) {
        adminUserRoleMapper.deleteByAdminUserId(adminUserId);
        for (Integer roleId : roleIds) {
            AdminUserRole adminUserRole = new AdminUserRole();
            adminUserRole.setAdminUserId(adminUserId);
            adminUserRole.setRoleId(roleId);
            adminUserRoleMapper.insert(adminUserRole);
        }
        return true;
    }

    @Override
    public List<Role> getAdminUserRoles(Integer adminUserId) {
        return adminUserRoleMapper.selectRolesByAdminUserId(adminUserId);
    }

    @Override
    public Result<?> updateAdminUser(Integer id, String username, Integer adminLevel, String newPassword) {
        try {
            AdminUser adminUser = adminUserMapper.selectById(id);
            if (adminUser == null) {
                return Result.error("管理员不存在");
            }

            if (id == 1) {
                return Result.error("不能修改超级管理员的信息");
            }

            if (username != null && !username.isEmpty()) {
                AdminUser existing = adminUserMapper.selectByUsername(username);
                if (existing != null && !existing.getId().equals(id)) {
                    return Result.error("用户名已存在");
                }
                adminUser.setUsername(username);
            }

            if (adminLevel != null) {
                adminUser.setAdminLevel(adminLevel);
            }

            adminUserMapper.updateById(adminUser);

            if (newPassword != null && !newPassword.isEmpty()) {
                String encryptedPassword = passwordEncoder.encode(newPassword);
                adminUserMapper.updatePassword(id, encryptedPassword);
            }

            return Result.success("修改成功", null);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}