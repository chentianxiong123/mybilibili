package com.mybilibili.common.entity;

import lombok.Data;
import java.util.Date;

@Data
public class User {
    private Integer id;
    private String username;
    private String password;
    private String nickname;
    private String avatar;
    private String email;
    private String phone;
    private Integer gender;
    private Date birthdate;
    private String signature;
    private Integer level;
    private Integer followingCount;
    private Integer followerCount;
    private Integer videoCount;
    private Integer likedCount;
    private Integer coinCount;
    private Integer pointCount;
    private Integer experience;
    private String bio;
    private String announcement;
    private Integer status;
    private Date createdAt;
    private Date updatedAt;
}