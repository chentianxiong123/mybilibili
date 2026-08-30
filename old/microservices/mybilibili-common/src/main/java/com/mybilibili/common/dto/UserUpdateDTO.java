package com.mybilibili.common.dto;

import jakarta.validation.constraints.Size;
import lombok.Data;

@Data
public class UserUpdateDTO {
    @Size(max = 20, message = "昵称最多20个字符")
    private String nickname;

    private String avatar;

    private String email;

    private String phone;

    private Integer gender;

    private String birthdate;

    @Size(max = 100, message = "签名最多100个字符")
    private String signature;

    @Size(max = 200, message = "简介最多200个字符")
    private String bio;

    @Size(max = 500, message = "公告最多500个字符")
    private String announcement;
}
