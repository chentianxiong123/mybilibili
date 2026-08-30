package com.mybilibili.common.vo;

import lombok.Data;

import java.util.Date;
import java.util.List;

@Data
public class ManuscriptVO {
    private Integer id;
    private String title;
    private String description;
    private String coverUrl;
    private Integer userId;
    private Integer categoryId;
    private String categoryName;

    // 统计字段
    private Integer viewCount;
    private Integer likeCount;
    private Integer coinCount;
    private Integer collectCount;
    private Integer shareCount;
    private Integer commentCount;
    private Integer danmakuCount;

    // 时长字段
    private String duration;           // 总时长显示字符串
    private Integer durationSeconds;   // 总时长秒数

    // 稿件状态
    private Integer status;
    private Integer reviewStatus;
    private String reviewReason;
    private Date reviewTime;
    private Integer reviewerId;

    // 时间戳
    private Date uploadTime;
    private Date updatedAt;

    // 关联数据
    private List<VideoItemVO> videos;   // 稿件包含的视频列表
    private UserInfo uploader;          // 上传者信息
    private List<String> tags;          // 标签列表

    // 首页展示用 - 第一个视频的信息
    private Integer firstVideoId;       // 第一个视频ID（用于跳转）
    private String firstVideoPlayUrl;   // 第一个视频播放地址

    @Data
    public static class VideoItemVO {
        private Integer id;
        private String title;
        private String description;
        private String playUrl;
        private String playUrlHd;
        private String playUrlSd;
        private String playUrlLd;
        private String duration;           // 时长显示字符串
        private Integer durationSeconds;   // 时长秒数
        private String sourceVideoUrl;
        private Integer videoOrder;
        private Integer status;
        private Integer processStatus;
        private Integer processProgress;
        private String processStage;
        private String processError;
    }

    @Data
    public static class UserInfo {
        private Integer id;
        private String name;
        private String avatar;
        private Integer level;
        private String bio;
        private String signature;
        private Integer followerCount;
        private Integer followingCount;
        private Integer likedCount;
        private Boolean following;
    }
}
