package com.mybilibili.user.controller;

import com.mybilibili.common.dto.AdminLoginDTO;
import jakarta.validation.Valid;
import com.mybilibili.common.entity.AdminUser;
import com.mybilibili.common.entity.Role;
import com.mybilibili.common.vo.Result;
import com.mybilibili.user.entity.AuditLog;
import com.mybilibili.user.service.AuditLogService;
import com.mybilibili.user.service.AdminUserService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import jakarta.servlet.http.HttpServletRequest;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/admin")
@Tag(name = "管理员接口", description = "管理员登录、注册和角色管理")
public class AdminUserController {

    @Autowired
    private AdminUserService adminUserService;

    @Autowired
    private AuditLogService auditLogService;

    @PostMapping("/login")
    @Operation(summary = "管理员登录", description = "管理员用户名密码登录，返回JWT令牌")
    public Result<?> login(@Valid @RequestBody AdminLoginDTO adminLoginDTO, HttpServletRequest request) {
        Result<?> result = adminUserService.login(adminLoginDTO);
        recordLoginAudit(request, adminLoginDTO.getUsername(), result);
        return result;
    }

    @PostMapping("/register")
    @Operation(summary = "管理员注册", description = "仅超级管理员可以创建新的管理员账号")
    public Result<?> register(
            @RequestBody AdminLoginDTO adminLoginDTO,
            @RequestHeader("X-User-Id") Integer operatorId,
            @RequestHeader("X-User-Role") String operatorRole,
            HttpServletRequest request) {
        if (!"超级管理员".equals(operatorRole)) {
            Result<?> result = Result.error("仅超级管理员可以创建新管理员");
            recordAudit(request, "admin", "admin_register", "admin_user", adminLoginDTO.getUsername(), result);
            return result;
        }
        Result<?> result = adminUserService.register(adminLoginDTO);
        recordAudit(request, "admin", "admin_register", "admin_user", adminLoginDTO.getUsername(), result);
        return result;
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
            @RequestBody List<Integer> roleIds,
            HttpServletRequest request) {
        try {
            boolean result = adminUserService.updateAdminUserRoles(id, roleIds);
            recordAudit(request, "role", "admin_role_assign", "admin_user", String.valueOf(id),
                    result ? Result.success("设置成功") : Result.error("设置失败"));
            if (result) {
                return Result.success("设置成功");
            } else {
                return Result.error("设置失败");
            }
        } catch (Exception e) {
            recordAuditFailure(request, "role", "admin_role_assign", "admin_user", String.valueOf(id), e.getMessage());
            return Result.error(e.getMessage());
        }
    }

    @PutMapping("/{id}")
    @Operation(summary = "修改管理员信息", description = "仅超级管理员可以修改其他管理员的信息")
    public Result<?> updateAdminUser(
            @Parameter(description = "管理员ID") @PathVariable Integer id,
            @RequestBody Map<String, Object> body,
            @RequestHeader("X-User-Id") Integer operatorId,
            @RequestHeader("X-User-Role") String operatorRole,
            HttpServletRequest request) {
        if (!"超级管理员".equals(operatorRole)) {
            Result<?> result = Result.error("仅超级管理员可以修改管理员信息");
            recordAudit(request, "admin", "admin_update", "admin_user", String.valueOf(id), result);
            return result;
        }
        String username = (String) body.get("username");
        Integer adminLevel = body.get("adminLevel") != null ? ((Number) body.get("adminLevel")).intValue() : null;
        String newPassword = (String) body.get("newPassword");
        Result<?> result = adminUserService.updateAdminUser(id, username, adminLevel, newPassword);
        recordAudit(request, "admin", "admin_update", "admin_user", String.valueOf(id), result);
        return result;
    }

    private void recordAudit(HttpServletRequest request, String module, String action, String targetType,
                             String targetId, Result<?> result) {
        auditLogService.recordFromRequest(request, module, action, targetType, targetId,
                result != null && result.getCode() != null && result.getCode() == 200 ? 1 : 0,
                result == null ? null : result.getMessage(),
                result == null || result.getData() == null ? null : result.getData().toString());
    }

    private void recordLoginAudit(HttpServletRequest request, String username, Result<?> result) {
        AuditLog auditLog = new AuditLog();
        auditLog.setOperatorName(username);
        auditLog.setOperatorId(extractAdminId(result));
        auditLog.setOperatorRole(extractAdminRole(result));
        auditLog.setModule("auth");
        auditLog.setAction("admin_login");
        auditLog.setTargetType("admin_user");
        auditLog.setTargetId(username);
        auditLog.setRequestMethod(request == null ? null : request.getMethod());
        auditLog.setRequestUri(request == null ? null : request.getRequestURI());
        auditLog.setClientIp(request == null ? null : request.getRemoteAddr());
        auditLog.setUserAgent(request == null ? null : request.getHeader("User-Agent"));
        auditLog.setResult(result != null && result.getCode() != null && result.getCode() == 200 ? 1 : 0);
        auditLog.setMessage(result == null ? null : result.getMessage());
        auditLog.setDetail(result == null || result.getData() == null ? null : result.getData().toString());
        auditLogService.record(auditLog);
    }

    private Integer extractAdminId(Result<?> result) {
        if (result == null || result.getData() == null) {
            return null;
        }
        Object data = result.getData();
        if (data instanceof Map<?, ?> map) {
            Object adminUser = map.get("adminUser");
            if (adminUser instanceof AdminUser user) {
                return user.getId();
            }
            Object tokenUser = map.get("user");
            if (tokenUser instanceof AdminUser user) {
                return user.getId();
            }
        }
        return null;
    }

    private String extractAdminRole(Result<?> result) {
        if (result == null || result.getData() == null) {
            return null;
        }
        Object data = result.getData();
        if (data instanceof Map<?, ?> map) {
            Object role = map.get("role");
            return role == null ? null : role.toString();
        }
        return null;
    }

    private void recordAuditFailure(HttpServletRequest request, String module, String action, String targetType,
                                    String targetId, String message) {
        auditLogService.recordFromRequest(request, module, action, targetType, targetId, 0, message, null);
    }
}
