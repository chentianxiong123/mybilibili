package com.mybilibili.common.entity;

import lombok.Data;

import java.util.Date;

@Data
public class CommentReply {
    private Integer id;
    private Integer commentId; // 被回复的评论ID
    private Integer userId; // 回复者ID
    private Integer replyUserId; // 被回复者ID
    private String content; // 回复内容
    private Integer likeCount; // 点赞数
    private Date createdAt;
    private Integer status; // 状态：0-正常，1-审核中，2-已删除
}
