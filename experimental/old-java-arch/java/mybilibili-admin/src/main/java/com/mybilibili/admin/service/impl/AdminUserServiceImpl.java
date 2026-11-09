package com.mybilibili.admin.service.impl;

import com.mybilibili.admin.dto.AdminLoginDTO;
import com.mybilibili.admin.entity.AdminUser;
import com.mybilibili.admin.mapper.AdminUserMapper;
import com.mybilibili.admin.mapper.AdminUserRoleMapper;
import com.mybilibili.admin.service.AdminUserService;
import com.mybilibili.common.entity.AdminUserRole;
import com.mybilibili.common.entity.Role;
import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.vo.Result;
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
            // 检查用户名是否已存在
            AdminUser existingAdmin = adminUserMapper.selectByUsername(adminLoginDTO.getUsername());
            if (existingAdmin != null) {
                return Result.error("用户名已存在");
            }

            // 对密码进行加密
            String encryptedPassword = passwordEncoder.encode(adminLoginDTO.getPassword());

            // 创建AdminUser对象
            AdminUser adminUser = new AdminUser();
            adminUser.setUsername(adminLoginDTO.getUsername());
            adminUser.setPassword(encryptedPassword);
            adminUser.setAdminLevel(1); // 默认普通管理员

            // 插入数据
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
            
            // 获取角色信息
            List<Role> roles = adminUserRoleMapper.selectRolesByAdminUserId(adminUser.getId());
            adminMap.put("roles", roles);
            
            // 生成角色名称字符串
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
        // 先删除原有的角色关联
        adminUserRoleMapper.deleteByAdminUserId(adminUserId);
        // 添加新的角色关联
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
}