package com.mybilibili.common.vo;

import lombok.Data;

import java.util.Date;

@Data
public class MessageSettingVO {
    private Integer id;
    private Integer userId;
    private Boolean privateMessageNotify;
    private Boolean replyNotify;
    private Boolean atNotify;
    private Boolean likeNotify;
    private Boolean systemNotify;
    private Date createdAt;
    private Date updatedAt;
}
