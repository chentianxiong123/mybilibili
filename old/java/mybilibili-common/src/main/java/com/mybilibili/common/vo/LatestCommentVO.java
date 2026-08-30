package com.mybilibili.common.vo;

import lombok.Data;

import java.util.Date;

/**
 * 最新评论VO
 */
@Data
public class LatestCommentVO {
    /**
     * 评论ID
     */
    private Integer id;

    /**
     * 评论内容
     */
    private String content;

    /**
     * 评论时间
     */
    private Date createTime;

    /**
     * 点赞数
     */
    private Integer likeCount;

    /**
     * 回复数
     */
    private Integer replyCount;

    /**
     * 评论者ID
     */
    private Integer userId;

    /**
     * 评论者昵称
     */
    private String userName;

    /**
     * 评论者头像
     */
    private String userAvatar;

    /**
     * 稿件ID
     */
    private Integer manuscriptId;

    /**
     * 稿件标题
     */
    private String manuscriptTitle;

    /**
     * 稿件封面
     */
    private String manuscriptCover;
}
