package com.mybilibili.user.controller;

import com.mybilibili.common.entity.Permission;
import com.mybilibili.common.entity.Role;
import com.mybilibili.common.vo.Result;
import com.mybilibili.user.mapper.PermissionMapper;
import com.mybilibili.user.mapper.RoleMapper;
import com.mybilibili.user.mapper.RolePermissionMapper;
import com.mybilibili.user.service.AuditLogService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.Date;
import java.util.List;

@RestController
@RequestMapping("/admin/roles")
@Tag(name = "角色权限管理", description = "角色CRUD、权限查询、角色权限分配")
public class RoleAdminController {

    @Autowired
    private RoleMapper roleMapper;

    @Autowired
    private PermissionMapper permissionMapper;

    @Autowired
    private RolePermissionMapper rolePermissionMapper;

    @Autowired
    private AuditLogService auditLogService;

    @GetMapping
    @Operation(summary = "获取角色列表")
    public Result<List<Role>> list() {
        return Result.success(roleMapper.selectList(null));
    }

    @GetMapping("/{id}")
    @Operation(summary = "获取角色详情")
    public Result<Role> getById(@PathVariable Integer id) {
        Role role = roleMapper.selectById(id);
        return role != null ? Result.success(role) : Result.error("角色不存在");
    }

    @PostMapping
    @Operation(summary = "添加角色")
    public Result<Role> create(@RequestBody Role role, jakarta.servlet.http.HttpServletRequest request) {
        if (role.getName() == null || role.getName().isEmpty()) {
            return Result.error("角色名称不能为空");
        }
        role.setCreateTime(new Date());
        role.setUpdateTime(new Date());
        roleMapper.insert(role);
        Result<Role> result = Result.success("添加成功", role);
        recordAudit(request, "role", "role_create", "role", String.valueOf(role.getId()), result);
        return result;
    }

    @PutMapping("/{id}")
    @Operation(summary = "编辑角色")
    public Result<Role> update(@PathVariable Integer id, @RequestBody Role role, jakarta.servlet.http.HttpServletRequest request) {
        Role existing = roleMapper.selectById(id);
        if (existing == null) {
            Result<Role> result = Result.error("角色不存在");
            recordAudit(request, "role", "role_update", "role", String.valueOf(id), result);
            return result;
        }
        existing.setName(role.getName());
        existing.setDescription(role.getDescription());
        existing.setUpdateTime(new Date());
        roleMapper.updateById(existing);
        Result<Role> result = Result.success("更新成功", existing);
        recordAudit(request, "role", "role_update", "role", String.valueOf(id), result);
        return result;
    }

    @DeleteMapping("/{id}")
    @Operation(summary = "删除角色")
    public Result<Void> delete(@PathVariable Integer id, jakarta.servlet.http.HttpServletRequest request) {
        rolePermissionMapper.deleteByRoleId(id);
        roleMapper.deleteById(id);
        Result<Void> result = Result.success("删除成功", null);
        recordAudit(request, "role", "role_delete", "role", String.valueOf(id), result);
        return result;
    }

    @GetMapping("/{id}/permissions")
    @Operation(summary = "获取角色权限")
    public Result<List<Permission>> getRolePermissions(@PathVariable Integer id) {
        return Result.success(rolePermissionMapper.selectPermissionsByRoleId(id));
    }

    @PutMapping("/{id}/permissions")
    @Operation(summary = "设置角色权限")
    public Result<Void> setRolePermissions(@PathVariable Integer id, @RequestBody List<Integer> permissionIds,
                                           jakarta.servlet.http.HttpServletRequest request) {
        rolePermissionMapper.deleteByRoleId(id);
        if (permissionIds != null) {
            for (Integer permId : permissionIds) {
                rolePermissionMapper.insert(id, permId);
            }
        }
        Result<Void> result = Result.success("权限设置成功", null);
        recordAudit(request, "role", "role_permission_update", "role", String.valueOf(id), result);
        return result;
    }

    @GetMapping("/permissions/all")
    @Operation(summary = "获取所有权限")
    public Result<List<Permission>> getAllPermissions() {
        return Result.success(permissionMapper.selectList(null));
    }

    private void recordAudit(jakarta.servlet.http.HttpServletRequest request, String module, String action,
                             String targetType, String targetId, Result<?> result) {
        auditLogService.recordFromRequest(request, module, action, targetType, targetId,
                result != null && result.getCode() != null && result.getCode() == 200 ? 1 : 0,
                result == null ? null : result.getMessage(),
                result == null || result.getData() == null ? null : result.getData().toString());
    }
}
