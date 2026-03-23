package com.mybilibili.common.entity;

import com.mybilibili.common.enums.TargetType;
import lombok.Data;

import java.util.Date;

@Data
public class Comment {
    private Integer id;
    private Integer manuscriptId;  // 关联稿件ID（向后兼容）
    private Integer userId;
    private String content;
    private Integer likeCount;
    private Integer replyCount;
    private String status;
    private Date createdAt;
    private Date updatedAt;

    // 新增字段：支持多种目标类型的评论
    private TargetType targetType;  // 评论目标类型：VIDEO-视频/DYNAMIC-动态
    private Integer targetId;       // 评论目标ID（根据targetType区分是manuscript_id还是dynamic_id）
}
