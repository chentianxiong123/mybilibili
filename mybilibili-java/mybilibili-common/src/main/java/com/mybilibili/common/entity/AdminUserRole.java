package com.mybilibili.common.entity;

public class AdminUserRole {
    private Integer adminUserId;
    private Integer roleId;

    // getters and setters
    public Integer getAdminUserId() {
        return adminUserId;
    }

    public void setAdminUserId(Integer adminUserId) {
        this.adminUserId = adminUserId;
    }

    public Integer getRoleId() {
        return roleId;
    }

    public void setRoleId(Integer roleId) {
        this.roleId = roleId;
    }
}