package com.mybilibili.common.vo;

import lombok.Data;

import java.util.Date;
import java.util.List;

/**
 * 视频搜索结果VO
 */
@Data
public class VideoSearchVO {

    /**
     * 视频ID
     */
    private Integer videoId;

    /**
     * 稿件ID
     */
    private Integer manuscriptId;

    /**
     * 视频标题
     */
    private String title;

    /**
     * 高亮标题（包含搜索关键词高亮）
     */
    private String highlightTitle;

    /**
     * 视频描述
     */
    private String description;

    /**
     * 高亮描述（包含搜索关键词高亮）
     */
    private String highlightDescription;

    /**
     * 封面URL
     */
    private String coverUrl;

    /**
     * 用户ID
     */
    private Integer userId;

    /**
     * 用户名
     */
    private String userName;

    /**
     * 用户头像
     */
    private String userAvatar;

    /**
     * 分类ID
     */
    private Integer categoryId;

    /**
     * 分类名称
     */
    private String categoryName;

    /**
     * 标签列表
     */
    private List<String> tags;

    /**
     * 播放量
     */
    private Integer viewCount;

    /**
     * 点赞数
     */
    private Integer likeCount;

    /**
     * 评论数
     */
    private Integer commentCount;

    /**
     * 分享数
     */
    private Integer shareCount;

    /**
     * 收藏数
     */
    private Integer collectCount;

    /**
     * 投币数
     */
    private Integer coinCount;

    /**
     * 视频时长（秒）
     */
    private Integer durationSeconds;

    /**
     * 视频时长（格式化字符串，如 05:30）
     */
    private String duration;

    /**
     * 上传时间
     */
    private Date uploadTime;

    /**
     * 视频状态
     */
    private Integer status;

    /**
     * 视频ID列表（稿件下的所有视频）
     */
    private List<Integer> videoIds;

    /**
     * 视频数量（分P数）
     */
    private Integer videoCount;
}
