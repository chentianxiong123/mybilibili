package com.mybilibili.common.entity;

import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.util.Date;

@Data
@TableName("comments")
public class Comment {
    private Integer id;
    private Integer manuscriptId;
    private Integer userId;
    private String content;
    private Integer likeCount;
    private Integer replyCount;
    private Date createdAt;
    private Date updatedAt;
    private Integer status;  // 0-正常 1-已删除
}
