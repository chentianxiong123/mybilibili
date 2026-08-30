package com.mybilibili.common.vo;

import lombok.Data;

@Data
public class UserVO {
    private Integer id;
    private String username;
    private String nickname;
    private String avatar;
    private Integer level;
    private Integer followingCount;
    private Integer followerCount;
    private Integer dynamicCount;
    private Integer videoCount;
    private Integer experience;
    private String bio;
    private String signature;
    private String announcement;
    private String email;
    private String phone;
    private Integer gender;
    private String birthdate;
    private Integer status;

    /**
     * 用户所有稿件的总播放数
     */
    private Integer totalViewCount;

    /**
     * 用户所有稿件的总获赞数
     */
    private Integer totalLikeCount;
}