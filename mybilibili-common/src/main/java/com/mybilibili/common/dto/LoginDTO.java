package com.mybilibili.common.dto;

import lombok.Data;

@Data
public class LoginDTO {
    private String username;
    private String password;
    private String email;
    private String emailCode;
    // 登录方式: "password" 或 "email_code"
    private String loginType;
    // 登录IP（用于日志记录）
    private String loginIp;
    private String userAgent;
}