package com.mybilibili.admin.controller;

import com.mybilibili.admin.service.RoleService;
import com.mybilibili.admin.service.PermissionService;
import com.mybilibili.common.entity.Role;
import com.mybilibili.common.entity.Permission;
import com.mybilibili.common.vo.Result;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/roles")
@Tag(name = "角色管理接口", description = "角色的增删改查和权限管理")
public class RoleController {

    @Autowired
    private RoleService roleService;

    @Autowired
    private PermissionService permissionService;

    @GetMapping
    @Operation(summary = "获取角色列表", description = "获取所有角色的列表")
    public Result<List<Role>> getRoleList() {
        try {
            List<Role> roles = roleService.getRoleList();
            return Result.success("获取成功", roles);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/{id}")
    @Operation(summary = "获取角色详情", description = "根据ID获取角色的详细信息")
    public Result<Role> getRoleById(@PathVariable Integer id) {
        try {
            Role role = roleService.getRoleById(id);
            if (role != null) {
                return Result.success("获取成功", role);
            } else {
                return Result.error("角色不存在");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping
    @Operation(summary = "添加角色", description = "添加新的角色")
    public Result<?> addRole(@RequestBody Role role) {
        try {
            boolean result = roleService.addRole(role);
            if (result) {
                return Result.success("添加成功");
            } else {
                return Result.error("添加失败");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PutMapping("/{id}")
    @Operation(summary = "更新角色", description = "更新角色的信息")
    public Result<?> updateRole(@PathVariable Integer id, @RequestBody Role role) {
        try {
            role.setId(id);
            boolean result = roleService.updateRole(role);
            if (result) {
                return Result.success("更新成功");
            } else {
                return Result.error("更新失败");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/{id}")
    @Operation(summary = "删除角色", description = "删除指定的角色")
    public Result<?> deleteRole(@PathVariable Integer id) {
        try {
            boolean result = roleService.deleteRole(id);
            if (result) {
                return Result.success("删除成功");
            } else {
                return Result.error("删除失败");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/{id}/permissions")
    @Operation(summary = "获取角色权限", description = "获取指定角色的权限列表")
    public Result<List<Permission>> getRolePermissions(@PathVariable Integer id) {
        try {
            List<Permission> permissions = roleService.getRolePermissions(id);
            return Result.success("获取成功", permissions);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PutMapping("/{id}/permissions")
    @Operation(summary = "设置角色权限", description = "为指定角色设置权限")
    public Result<?> setRolePermissions(@PathVariable Integer id, @RequestBody List<Integer> permissionIds) {
        try {
            boolean result = roleService.setRolePermissions(id, permissionIds);
            if (result) {
                return Result.success("设置成功");
            } else {
                return Result.error("设置失败");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/permissions/all")
    @Operation(summary = "获取所有权限", description = "获取系统中所有的权限")
    public Result<List<Permission>> getAllPermissions() {
        try {
            List<Permission> permissions = permissionService.getPermissionList();
            return Result.success("获取成功", permissions);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}