package com.mybilibili.common.entity;

import lombok.Data;

import java.util.Date;

@Data
public class MessageSetting {
    private Long id;
    private Integer userId;
    private Boolean privateMessageNotify;
    private Boolean replyNotify;
    private Boolean atNotify;
    private Boolean likeNotify;
    private Boolean systemNotify;
    private Date createdAt;
    private Date updatedAt;
}
