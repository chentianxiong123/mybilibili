package com.mybilibili.user.controller;

import com.mybilibili.common.dto.AdminLoginDTO;
import com.mybilibili.common.entity.AdminUser;
import com.mybilibili.common.entity.Role;
import com.mybilibili.common.vo.Result;
import com.mybilibili.user.service.AdminUserService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/admin")
@Tag(name = "管理员接口", description = "管理员登录、注册和角色管理")
public class AdminUserController {

    @Autowired
    private AdminUserService adminUserService;

    @PostMapping("/login")
    @Operation(summary = "管理员登录", description = "管理员用户名密码登录，返回JWT令牌")
    public Result<?> login(@RequestBody AdminLoginDTO adminLoginDTO) {
        return adminUserService.login(adminLoginDTO);
    }

    @PostMapping("/register")
    @Operation(summary = "管理员注册", description = "仅超级管理员(ID=1)可以创建新的管理员账号")
    public Result<?> register(
            @RequestBody AdminLoginDTO adminLoginDTO,
            @RequestHeader("X-User-Id") Integer operatorId) {
        if (operatorId == null || operatorId != 1) {
            return Result.error("仅超级管理员可以创建新管理员");
        }
        return adminUserService.register(adminLoginDTO);
    }

    @GetMapping("/list")
    @Operation(summary = "获取管理员列表", description = "获取所有管理员用户的列表")
    public Result<List<Map<String, Object>>> getAdminUserList() {
        try {
            List<Map<String, Object>> adminUsers = adminUserService.getAdminUserList();
            return Result.success("获取成功", adminUsers);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/{id}")
    @Operation(summary = "获取管理员详情", description = "根据ID获取管理员用户的详细信息")
    public Result<AdminUser> getAdminUserById(
            @Parameter(description = "管理员ID") @PathVariable Integer id) {
        try {
            AdminUser adminUser = adminUserService.getAdminUserById(id);
            if (adminUser != null) {
                return Result.success("获取成功", adminUser);
            } else {
                return Result.error("管理员用户不存在");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/{id}/roles")
    @Operation(summary = "获取管理员角色", description = "获取指定管理员用户的角色列表")
    public Result<List<Role>> getAdminUserRoles(
            @Parameter(description = "管理员ID") @PathVariable Integer id) {
        try {
            List<Role> roles = adminUserService.getAdminUserRoles(id);
            return Result.success("获取成功", roles);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PutMapping("/{id}/roles")
    @Operation(summary = "设置管理员角色", description = "为指定管理员用户分配角色")
    public Result<?> setAdminUserRoles(
            @Parameter(description = "管理员ID") @PathVariable Integer id,
            @RequestBody List<Integer> roleIds) {
        try {
            boolean result = adminUserService.updateAdminUserRoles(id, roleIds);
            if (result) {
                return Result.success("设置成功");
            } else {
                return Result.error("设置失败");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PutMapping("/{id}")
    @Operation(summary = "修改管理员信息", description = "仅超级管理员(ID=1)可以修改其他管理员的信息")
    public Result<?> updateAdminUser(
            @Parameter(description = "管理员ID") @PathVariable Integer id,
            @RequestBody Map<String, Object> body,
            @RequestHeader("X-User-Id") Integer operatorId) {
        if (operatorId == null || operatorId != 1) {
            return Result.error("仅超级管理员可以修改管理员信息");
        }
        String username = (String) body.get("username");
        Integer adminLevel = body.get("adminLevel") != null ? ((Number) body.get("adminLevel")).intValue() : null;
        String newPassword = (String) body.get("newPassword");
        return adminUserService.updateAdminUser(id, username, adminLevel, newPassword);
    }
}