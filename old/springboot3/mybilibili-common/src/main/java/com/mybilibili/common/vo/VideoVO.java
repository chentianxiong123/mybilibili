package com.mybilibili.common.vo;

import lombok.Data;

import java.util.Date;
import java.util.List;

@Data
public class VideoVO {
    private Integer id;
    private String title;
    private String description;
    private String coverUrl;
    private String playUrl;
    private String playUrlHd;
    private String playUrlSd;
    private String playUrlLd;
    private UserInfo uploader;
    private Integer categoryId;
    private String categoryName;
    private List<String> tags;
    private Integer viewCount;
    private Integer likeCount;
    private Integer coinCount;
    private Integer collectCount;
    private Integer shareCount;
    private Integer commentCount;
    private Integer danmakuCount;
    private String duration;
    private Integer status;
    private Date uploadTime;

    // 视频顺序（分P）
    private Integer videoOrder;

    // 稿件关联信息
    private Integer manuscriptId;
    private String manuscriptTitle;  // 稿件标题（用于列表展示）
    private ManuscriptInfo manuscript;

    // 分P相关字段（用于稿件详情页）
    private List<VideoItemVO> manuscriptVideos;  // 稿件下的所有视频列表
    private Integer currentVideoIndex;           // 当前播放的是第几个视频（从0开始）
    private Integer totalVideos;                 // 稿件下视频总数

    @Data
    public static class UserInfo {
        private Integer id;
        private String name;
        private String avatar;
        private Integer level;
        private String bio;
        private String signature;
        private Integer followerCount;
        private Boolean following;
    }

    @Data
    public static class ManuscriptInfo {
        private Integer id;
        private String title;
        private String description;
        private String coverUrl;
        private Integer status;
        private Integer reviewStatus;
        private Date uploadTime;
    }

    @Data
    public static class VideoItemVO {
        private Integer id;
        private String title;
        private Integer videoOrder;  // 视频顺序
        private String duration;
        private String playUrl;
    }
}
