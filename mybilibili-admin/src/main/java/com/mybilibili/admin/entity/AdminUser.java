package com.mybilibili.admin.entity;

import lombok.Data;

import java.util.Date;

@Data
public class AdminUser {
    private Integer id;
    private String username;
    private String password;
    private Integer adminLevel;
    private Date createdAt;
    private Date updatedAt;
}