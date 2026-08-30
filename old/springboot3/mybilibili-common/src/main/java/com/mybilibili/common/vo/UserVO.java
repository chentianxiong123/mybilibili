package com.mybilibili.common.vo;

import lombok.Data;

import java.util.List;

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
    private Integer manuscriptCount;
    private Integer experience;
    private String bio;
    private String signature;
    private String announcement;
    private String email;
    private String phone;
    private Integer gender;
    private String birthdate;
    private Integer status;
    private Integer coinCount;

    private Integer totalViewCount;

    private Integer totalLikeCount;

    private Integer pinnedVideoId;

    private ManuscriptVO pinnedVideo;

    private List<String> tags;
}