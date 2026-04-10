package com.mybilibili.common.entity;

import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.util.Date;

@Data
@TableName("replies")
public class Reply {
    private Integer id;
    private Integer commentId;
    private Integer userId;
    private Integer replyToUserId;
    private String content;
    private Integer likeCount;
    private String status;  // NORMAL-正常 REMOVED-已删除
    private Date createdAt;
    private Date updatedAt;
}