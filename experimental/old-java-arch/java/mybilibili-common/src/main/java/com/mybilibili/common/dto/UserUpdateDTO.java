package com.mybilibili.common.dto;

import lombok.Data;

@Data
public class UserUpdateDTO {
    private String nickname;
    private String avatar;
    private String email;
    private String phone;
    private Integer gender;
    private String birthdate;
    private String signature;
    private String bio;
    private String announcement;
}