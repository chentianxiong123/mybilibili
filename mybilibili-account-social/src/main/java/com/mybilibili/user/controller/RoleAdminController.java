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
import java.util.LinkedHashMap;
import java.util.Map;
import java.util.List;
import java.util.Set;

@RestController
@RequestMapping("/admin/roles")
@Tag(name = "角色权限管理", description = "角色CRUD、权限查询、角色权限分配")
public class RoleAdminController {

    private static final Map<String, RoleTemplate> ROLE_TEMPLATES = createRoleTemplates();

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

    @GetMapping("/templates")
    @Operation(summary = "获取岗位权限模板")
    public Result<List<RoleTemplate>> getRoleTemplates() {
        return Result.success(List.copyOf(ROLE_TEMPLATES.values()));
    }

    @PutMapping("/{id}/template/{templateCode}")
    @Operation(summary = "应用岗位权限模板")
    public Result<Void> applyRoleTemplate(@PathVariable Integer id,
                                          @PathVariable String templateCode,
                                          jakarta.servlet.http.HttpServletRequest request) {
        Role role = roleMapper.selectById(id);
        if (role == null) {
            Result<Void> result = Result.error("角色不存在");
            recordAudit(request, "role", "role_template_apply", "role", String.valueOf(id), result);
            return result;
        }
        RoleTemplate template = ROLE_TEMPLATES.get(templateCode);
        if (template == null) {
            Result<Void> result = Result.error("岗位模板不存在");
            recordAudit(request, "role", "role_template_apply", "role", String.valueOf(id), result);
            return result;
        }
        List<Permission> permissions = permissionMapper.selectByCodes(template.permissionCodes());
        Set<String> existingCodes = permissions.stream()
                .map(Permission::getCode)
                .collect(java.util.stream.Collectors.toSet());
        List<String> missingCodes = template.permissionCodes().stream()
                .filter(code -> !existingCodes.contains(code))
                .toList();
        if (!missingCodes.isEmpty()) {
            Result<Void> result = Result.error("权限码不存在: " + String.join(", ", missingCodes));
            recordAudit(request, "role", "role_template_apply", "role", String.valueOf(id), result);
            return result;
        }
        rolePermissionMapper.deleteByRoleId(id);
        for (Permission permission : permissions) {
            rolePermissionMapper.insert(id, permission.getId());
        }
        Result<Void> result = Result.success("岗位模板已应用", null);
        recordAudit(request, "role", "role_template_apply", "role", String.valueOf(id), result);
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

    private static Map<String, RoleTemplate> createRoleTemplates() {
        Map<String, RoleTemplate> templates = new LinkedHashMap<>();
        templates.put("platform-operation", new RoleTemplate(
                "platform-operation",
                "平台运营",
                "适合处理工单、运营任务、推荐策略、索引和运营审计",
                List.of("operation:manage", "search:manage", "audit:manage", "statistics:manage")
        ));
        templates.put("content-review", new RoleTemplate(
                "content-review",
                "内容审核",
                "适合处理稿件审核、内容审核、举报、评论和违禁词",
                List.of("review:manage", "comment:manage")
        ));
        templates.put("ai-manager", new RoleTemplate(
                "ai-manager",
                "AI 管理",
                "适合维护 AI 渠道、技能、用量和客服会话",
                List.of("ai:manage")
        ));
        templates.put("media-manager", new RoleTemplate(
                "media-manager",
                "媒体管理",
                "适合维护视频、字幕、分类、轮播图、直播和会议资源",
                List.of("video:manage", "category:manage", "banner:manage", "live:manage", "meeting:manage")
        ));
        templates.put("system-manager", new RoleTemplate(
                "system-manager",
                "系统管理",
                "适合维护管理员、角色权限和安全日志",
                List.of("admin:manage", "role:manage", "security:manage")
        ));
        templates.put("super-admin", new RoleTemplate(
                "super-admin",
                "超级管理员",
                "完整后台权限模板，仅用于初始化或修复超级管理员角色",
                List.of(
                        "user:manage",
                        "video:manage",
                        "comment:manage",
                        "category:manage",
                        "tag:manage",
                        "review:manage",
                        "statistics:manage",
                        "role:manage",
                        "admin:manage",
                        "security:manage",
                        "live:manage",
                        "meeting:manage",
                        "storage:manage",
                        "banner:manage",
                        "search:manage",
                        "ai:manage",
                        "message:manage",
                        "audit:manage",
                        "operation:manage"
                )
        ));
        return templates;
    }

    public record RoleTemplate(String code, String name, String description, List<String> permissionCodes) {
    }
}
