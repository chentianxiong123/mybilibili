package com.mybilibili.admin.controller;

import com.mybilibili.admin.service.UserService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/users")
@Tag(name = "用户管理接口", description = "用户管理相关操作")
public class UserController {

    @Autowired
    private UserService userService;

    @GetMapping
    @Operation(summary = "获取用户列表", description = "获取用户列表，支持分页和搜索")
    public Object getUserList(
            @RequestParam(defaultValue = "1") Integer page,
            @RequestParam(defaultValue = "10") Integer size,
            @RequestParam(required = false) String keyword) {
        return userService.getUserList(page, size, keyword);
    }

    @GetMapping("/{id}")
    @Operation(summary = "获取用户详情", description = "根据ID获取用户详情")
    public Object getUserById(@PathVariable Integer id) {
        return userService.getUserById(id);
    }

    @PutMapping("/{id}/status")
    @Operation(summary = "更新用户状态", description = "更新用户状态，例如禁用/启用")
    public Object updateUserStatus(@PathVariable Integer id, @RequestParam Integer status) {
        return userService.updateUserStatus(id, status);
    }

    @PutMapping("/{id}/password")
    @Operation(summary = "重置用户密码", description = "重置用户密码")
    public Object resetPassword(@PathVariable Integer id, @RequestParam String newPassword) {
        return userService.resetPassword(id, newPassword);
    }
}