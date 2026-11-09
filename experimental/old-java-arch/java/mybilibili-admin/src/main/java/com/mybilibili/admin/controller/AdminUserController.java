package com.mybilibili.admin.controller;

import com.mybilibili.admin.dto.AdminLoginDTO;
import com.mybilibili.admin.entity.AdminUser;
import com.mybilibili.admin.service.AdminUserService;
import com.mybilibili.common.entity.Role;
import com.mybilibili.common.vo.Result;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.security.SecurityRequirement;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/user")
@Tag(name = "管理员用户接口", description = "管理员登录、注册和角色管理")
public class AdminUserController {

    @Autowired
    private AdminUserService adminUserService;

    @PostMapping("/login")
    @Operation(summary = "管理员登录", description = "管理员登录接口")
    public Object login(@RequestBody AdminLoginDTO adminLoginDTO) {
        return adminUserService.login(adminLoginDTO);
    }

    @PostMapping("/register")
    @Operation(summary = "管理员注册", description = "管理员注册接口")
    public Object register(@RequestBody AdminLoginDTO adminLoginDTO) {
        return adminUserService.register(adminLoginDTO);
    }

    @GetMapping("/list")
    @Operation(summary = "获取管理员用户列表", description = "获取所有管理员用户的列表")
    @SecurityRequirement(name = "JWT")
    public Result<List<Map<String, Object>>> getAdminUserList() {
        try {
            List<Map<String, Object>> adminUsers = adminUserService.getAdminUserList();
            return Result.success("获取成功", adminUsers);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/{id}")
    @Operation(summary = "获取管理员用户详情", description = "根据ID获取管理员用户的详细信息")
    @SecurityRequirement(name = "JWT")
    public Result<AdminUser> getAdminUserById(@PathVariable Integer id) {
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
    @Operation(summary = "获取管理员用户角色", description = "获取指定管理员用户的角色列表")
    @SecurityRequirement(name = "JWT")
    public Result<List<Role>> getAdminUserRoles(@PathVariable Integer id) {
        try {
            List<Role> roles = adminUserService.getAdminUserRoles(id);
            return Result.success("获取成功", roles);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PutMapping("/{id}/roles")
    @Operation(summary = "设置管理员用户角色", description = "为指定管理员用户设置角色")
    @SecurityRequirement(name = "JWT")
    public Result<?> setAdminUserRoles(@PathVariable Integer id, @RequestBody List<Integer> roleIds) {
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
}