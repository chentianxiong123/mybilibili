package com.mybilibili.common.dto;

import jakarta.validation.constraints.NotBlank;
import lombok.Data;

@Data
public class LoginDTO {
    private String username;
    private String password;
    private String email;
    private String emailCode;

    @NotBlank(message = "登录方式不能为空")
    private String loginType;

    private String loginIp;
    private String userAgent;
}
