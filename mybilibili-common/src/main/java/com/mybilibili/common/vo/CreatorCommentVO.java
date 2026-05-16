package com.mybilibili.common.vo;

import lombok.Data;

import java.util.Date;

/**
 * 创作者评论视图对象
 * 用于创作者管理自己稿件下的评论
 */
@Data
public class CreatorCommentVO {
    private Integer id;
    private Integer manuscriptId;
    private String manuscriptTitle;
    private String manuscriptCover;
    private Integer userId;
    private String userName;
    private String userAvatar;
    private String content;
    private Integer likeCount;
    private Integer replyCount;
    private Date createTime;
    private boolean liked;
    private String commentType; // "comment" or "reply"
    private Integer parentCommentId; // for replies, the parent comment id
    private String replyToUserName; // for replies, who is being replied to
}
